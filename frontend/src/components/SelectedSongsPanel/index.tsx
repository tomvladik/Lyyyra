import { useContext, useMemo, useState } from "react";
import { GetCombinedPdf, GetSongProjection, GetSongVerses } from "../../../wailsjs/go/main/App";
import { SelectionContext } from "../../selectionContext";
import { PdfModal } from "../PdfModal";
import styles from "./index.module.less";

export const SelectedSongsPanel = () => {
    const { selectedSongs, removeSongFromSelection, clearSelection } = useContext(SelectionContext);
    const [isCombining, setIsCombining] = useState(false);
    const [combinedPdf, setCombinedPdf] = useState("");
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [error, setError] = useState("");

    const panelTitle = useMemo(() => {
        if (!selectedSongs.length) return "Výběr je prázdný";
        if (selectedSongs.length === 1) return "1 skladba ve výběru";
        return `skladeb ve výběru: ${selectedSongs.length}`;
    }, [selectedSongs.length]);

    const handleRemove = (id: number) => removeSongFromSelection(id);

    const handleCombineClick = async () => {
        if (!selectedSongs.length) return;
        setIsCombining(true);
        setError("");
        try {
            const filenames = selectedSongs.map(song => song.filename);
            const dataUrl = await GetCombinedPdf(filenames);
            setCombinedPdf(dataUrl);
            setIsModalOpen(true);
        } catch (err) {
            console.error("Failed to create combined PDF", err);
            setError("Nepodařilo se vytvořit společné PDF. Zkuste to prosím znovu.");
        } finally {
            setIsCombining(false);
        }
    };

    const handleProjectClick = async () => {
        if (!selectedSongs.length) return;
        // Open the window synchronously to avoid popup blockers.
        const w = window.open('', '_blank');
        if (!w) {
            setError('Nelze otevřít projekční okno. Zkontrolujte blokování vyskakovacích oken.');
            return;
        }

        try {
            const getProj = typeof GetSongProjection === 'function' ? GetSongProjection : undefined;
            const getVerses = typeof GetSongVerses === 'function' ? GetSongVerses : undefined;
            const songsData: Array<{ title: string; verseOrder: string; verses: Array<{ name: string; lines: string }> }> = [];

            for (const song of selectedSongs) {
                let payloadRaw = '';
                if (typeof getProj === 'function') {
                    payloadRaw = await getProj(song.id);
                } else if (typeof getVerses === 'function') {
                    const v = await getVerses(song.id);
                    const parts = (v || '').split('===').map((s: string) => s.trim()).filter(Boolean);
                    const verses = parts.map((lines: string, idx: number) => ({ name: `v${idx + 1}`, lines }));
                    songsData.push({ title: song.title, verseOrder: verses.map((_: any, i: number) => `v${i + 1}`).join(' '), verses });
                    continue;
                }

                if (payloadRaw) {
                    try {
                        const obj = JSON.parse(payloadRaw);
                        const verses = Array.isArray(obj.verses) ? obj.verses.map((vv: any) => ({ name: vv.name || '', lines: vv.lines || '' })) : [];
                        const verseOrder = typeof obj.verse_order === 'string' ? obj.verse_order : '';
                        songsData.push({ title: song.title, verseOrder, verses });
                    } catch (e) {
                        console.warn('Failed to parse projection payload, falling back', e);
                        songsData.push({ title: song.title, verseOrder: '', verses: [] });
                    }
                } else {
                    songsData.push({ title: song.title, verseOrder: '', verses: [] });
                }
            }

            // Safely serialize songs data for embedding inside a <script>.
            // Escape problematic characters: newlines, line/paragraph separators, and </script> tags.
            const safeSongsJson = encodeURIComponent(JSON.stringify(songsData)
                .replace(/\n/g, '\\n')
                .replace(/\r/g, '\\r')
                .replace(/\u2028/g, '\\u2028')
                .replace(/\u2029/g, '\\u2029')
                .replace(/<\/script>/g, '<\\/script>')
            );

            // Build the inline JavaScript code with properly escaped backslashes for template literal
            const inlineScript = `
    const songs = JSON.parse(decodeURIComponent('${safeSongsJson}'));
    let songIdx = 0;
    let verseIdx = 0;

    function parseOrder(orderStr, verses) {
    if (!orderStr || !orderStr.trim()) return verses.map(v=>v.name);
    return orderStr.split(/\\s+/).filter(Boolean);
    }

    function currentSequence() {
    const s = songs[songIdx];
    return parseOrder(s.verseOrder, s.verses);
    }

    function show() {
    const s = songs[songIdx];
    const seq = currentSequence();
    const name = seq[verseIdx] || '';
    const verseObj = s.verses.find(v=>v.name===name) || s.verses[verseIdx] || {lines: ''};
    const linesHtml = (verseObj.lines || '').split('\\n').map(function(l){ return '<div class="verse">' + l.replace(/</g,'&lt;') + '</div>'; }).join('');
    document.getElementById('title').textContent = s.title || '';
    document.getElementById('verseContainer').innerHTML = linesHtml;
    }

    function clampVerse() {
    const seq = currentSequence();
    if (verseIdx < 0) verseIdx = 0;
    if (verseIdx >= seq.length) verseIdx = seq.length - 1;
    }

    function prevVerse(){ verseIdx--; clampVerse(); show(); }
    function nextVerse(){ verseIdx++; clampVerse(); show(); }
    function prevSong(){ songIdx = Math.max(0, songIdx-1); verseIdx = 0; show(); }
    function nextSong(){ songIdx = Math.min(songs.length-1, songIdx+1); verseIdx = 0; show(); }

    document.getElementById('prevVerse').addEventListener('click', prevVerse);
    document.getElementById('nextVerse').addEventListener('click', nextVerse);
    document.getElementById('prevSong').addEventListener('click', prevSong);
    document.getElementById('nextSong').addEventListener('click', nextSong);
    document.getElementById('fullscreen').addEventListener('click', ()=>{
    const el = document.documentElement;
    if (!document.fullscreenElement) el.requestFullscreen && el.requestFullscreen(); else document.exitFullscreen && document.exitFullscreen();
    });

    document.addEventListener('keydown', (e)=>{
    if (e.key === 'ArrowLeft') prevVerse();
    if (e.key === 'ArrowRight') nextVerse();
    if (e.key === 'ArrowUp') prevSong();
    if (e.key === 'ArrowDown') nextSong();
    if (e.key === 'f' || e.key === 'F') document.documentElement.requestFullscreen && document.documentElement.requestFullscreen();
    });

    show();
  `;

            const html = `<!doctype html>
<html>
  <head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Projection</title>
  <style>
    html,body{height:100%;margin:0;background:#000;color:#fff;font-family:Arial, Helvetica, sans-serif}
    .container{display:flex;flex-direction:column;align-items:stretch;height:100%}
    .controls{display:flex;gap:8px;padding:12px;background:rgba(0,0,0,0.2);position:fixed;right:12px;top:12px;z-index:999}
    .btn{background:rgba(255,255,255,0.06);color:#fff;padding:8px 12px;border:1px solid rgba(255,255,255,0.06);border-radius:6px;cursor:pointer}
    .stage{flex:1;display:flex;flex-direction:column;align-items:center;justify-content:center;padding:40px}
    .title{font-size:48px;margin:0 0 16px 0;color:#fff;text-align:center}
    .verse{font-size:40px;line-height:1.4;margin:6px 0;color:#fff;text-align:center;max-width:1600px}
    .meta{position:fixed;left:12px;top:12px;color:rgba(255,255,255,0.6);font-size:14px}
    @media (min-width:1200px){.title{font-size:64px}.verse{font-size:48px}}
  </style>
  </head>
  <body>
  <div class="container">
    <div class="meta">Use ←/→ verses • ↑/↓ songs • F fullscreen</div>
    <div class="controls">
    <button class="btn" id="prevSong">◀︎ Song</button>
    <button class="btn" id="prevVerse">◀︎ Verse</button>
    <button class="btn" id="nextVerse">Verse ▶︎</button>
    <button class="btn" id="nextSong">Song ▶︎</button>
    <button class="btn" id="fullscreen">Fullscreen</button>
    </div>
    <div class="stage">
    <h1 class="title" id="title"></h1>
    <div id="verseContainer"></div>
    </div>
  </div>
  <script>${inlineScript}</script>
  </body>
</html>`;

            // Use a Blob URL to navigate the newly opened window to the generated HTML.
            // This avoids issues with unescaped characters when calling document.write.
            const blob = new Blob([html], { type: 'text/html;charset=utf-8' });
            const url = URL.createObjectURL(blob);
            try {
                w.location.href = url;
                // Revoke when the window unloads to free memory.
                w.addEventListener && w.addEventListener('beforeunload', () => URL.revokeObjectURL(url));
            } catch (e) {
                // Fallback: some environments may block navigation; try document.write as before.
                try {
                    w.document.open();
                    w.document.write(html);
                    w.document.close();
                } catch (e2) {
                    URL.revokeObjectURL(url);
                    throw e2;
                }
            }
        } catch (err) {
            console.error('Projection failed', err);
            setError('Nepodařilo se otevřít projekční okno.');
        }
    };

    const handleCloseModal = () => {
        setIsModalOpen(false);
        setCombinedPdf("");
    };

    if (!selectedSongs.length) {
        return null;
    }

    return (
        <aside className={styles.panel} aria-label="Výběr skladeb">
            <header className={styles.panelHeader}>
                <div>
                    <h1 className={styles.panelTitle}>Připravené noty</h1>
                    <p className={styles.panelLabel}>{panelTitle}</p>
                </div>
                <button
                    type="button"
                    className={styles.clearButton}
                    onClick={clearSelection}
                    disabled={!selectedSongs.length || isCombining}
                >Zrušit</button>
            </header>

            <div className={styles.list} role="list">
                {!selectedSongs.length && (
                    <p className={styles.emptyState}>
                        Klepněte na ikonu 📋 u libovolné skladby a vytvořte kolekci pro tisk.
                    </p>
                )}
                {selectedSongs.map(song => (
                    <div key={song.id} className={styles.listItem} role="listitem">
                        <div>
                            <span className={styles.itemNumber}>{song.entry}.</span>
                            <span className={styles.itemTitle}>{song.title}</span>
                        </div>
                        <button
                            type="button"
                            className={styles.removeButton}
                            onClick={() => handleRemove(song.id)}
                            title="Odebrat ze seznamu"
                        >
                            ✕
                        </button>
                    </div>
                ))}
            </div>

            <div className={styles.actions}>
                <button
                    type="button"
                    className={styles.actionButton}
                    onClick={handleCombineClick}
                    disabled={!selectedSongs.length || isCombining}
                >
                    {isCombining ? "Vytvářím PDF…" : "Zobrazit připravené noty"}
                </button>
                <button
                    type="button"
                    className={styles.actionButton}
                    onClick={handleProjectClick}
                    disabled={!selectedSongs.length || isCombining}
                >
                    Promítat texty
                </button>
                {error && <p className={styles.errorText}>{error}</p>}
            </div>

            <PdfModal
                isOpen={isModalOpen}
                dataUrl={combinedPdf}
                songName="Připravené noty"
                onClose={handleCloseModal}
            />
        </aside>
    );
};

export default SelectedSongsPanel;
