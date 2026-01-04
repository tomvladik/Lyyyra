import { useContext, useEffect, useMemo, useRef, useState } from "react";
import { GetCombinedPdf, GetSongProjection, GetSongVerses } from "../../../wailsjs/go/app/App";
import { SelectionContext } from "../../selectionContext";
import { PdfModal } from "../PdfModal";
import styles from "./index.module.less";
import projectionTemplate from "./projection-template.html?raw";

export const SelectedSongsPanel = () => {
    const { selectedSongs, removeSongFromSelection, clearSelection } = useContext(SelectionContext);
    const projectionWindowRef = useRef<Window | null>(null);
    const [isCombining, setIsCombining] = useState(false);
    const [combinedPdf, setCombinedPdf] = useState("");
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [error, setError] = useState("");
    const [isProjectionOpen, setIsProjectionOpen] = useState(false);
    const [projectionSongsData, setProjectionSongsData] = useState<Array<{ title: string; verseOrder: string; verses: Array<{ name: string; lines: string }> }>>([]);
    const [currentSongIdx, setCurrentSongIdx] = useState(0);
    const [currentVerseIdx, setCurrentVerseIdx] = useState(0);

    // Monitor projection window status periodically
    useEffect(() => {
        if (!isProjectionOpen) return;

        console.log('[Projection] Window monitoring started');

        const checkInterval = setInterval(() => {
            if (projectionWindowRef.current) {
                const isClosed = projectionWindowRef.current.closed;
                console.log('[Projection] Window check - closed:', isClosed);
                if (isClosed) {
                    console.log('[Projection] Window detected as closed, updating state');
                    setIsProjectionOpen(false);
                    projectionWindowRef.current = null;
                }
            }
        }, 500); // Check every 500ms

        return () => {
            console.log('[Projection] Window monitoring stopped');
            clearInterval(checkInterval);
        };
    }, [isProjectionOpen]);

    const sendProjectionCommand = (command: 'nextVerse' | 'prevVerse' | 'nextSong' | 'prevSong') => {
        if (projectionWindowRef.current && !projectionWindowRef.current.closed) {
            projectionWindowRef.current.postMessage({ type: 'projection-control', command }, '*');
        }
    };

    const closeProjection = () => {
        if (projectionWindowRef.current && !projectionWindowRef.current.closed) {
            projectionWindowRef.current.close();
        }
        setIsProjectionOpen(false);
        projectionWindowRef.current = null;
    };

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

        console.log('[Projection] handleProjectClick called');

        // Prevent opening multiple projection windows
        if (projectionWindowRef.current && !projectionWindowRef.current.closed) {
            console.log('[Projection] Window already open, preventing multiple opens');
            setError('Projekční okno je již otevřeno.');
            return;
        }

        // Open the window synchronously to avoid popup blockers.
        const w = window.open('', '_blank');
        if (!w) {
            console.error('[Projection] Failed to open window - popups blocked?');
            setError('Nelze otevřít projekční okno. Zkontrolujte blokování vyskakovacích oken.');
            return;
        }

        console.log('[Projection] Window opened successfully');

        try {
            // Store the window reference for control
            projectionWindowRef.current = w;

            // Set projection as open - triggers re-render for control panel
            console.log('[Projection] Setting isProjectionOpen = true');
            setIsProjectionOpen(true);
            setError("");
            setCurrentSongIdx(0);
            setCurrentVerseIdx(0);

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
                        songsData.push({ title: `${song.entry}: ${song.title}`, verseOrder, verses });
                    } catch (e) {
                        console.warn('Failed to parse projection payload, falling back', e);
                        songsData.push({ title: `${song.entry}: ${song.title}`, verseOrder: '', verses: [] });
                    }
                } else {
                    songsData.push({ title: `${song.entry}: ${song.title}`, verseOrder: '', verses: [] });
                }
            }

            // Store songs data for preview display
            setProjectionSongsData(songsData);

            // Listen for projection window state updates
            window.addEventListener('message', (event) => {
                if (event.data && event.data.type === 'projection-state') {
                    setCurrentSongIdx(event.data.songIdx || 0);
                    setCurrentVerseIdx(event.data.verseIdx || 0);
                }
            });

            // Safely serialize songs data for embedding inside a <script>.
            // Escape problematic characters: newlines, line/paragraph separators, and </script> tags.
            const safeSongsJson = encodeURIComponent(JSON.stringify(songsData)
                .replace(/\n/g, '\\n')
                .replace(/\r/g, '\\r')
                .replace(/\u2028/g, '\\u2028')
                .replace(/\u2029/g, '\\u2029')
                .replace(/<\/script>/g, '<\\/script>')
            );

            const html = projectionTemplate.replace('{{SONGS_DATA}}', safeSongsJson);

            // Use a Blob URL to navigate the newly opened window to the generated HTML.
            // This avoids issues with unescaped characters when calling document.write.
            const blob = new Blob([html], { type: 'text/html;charset=utf-8' });
            const url = URL.createObjectURL(blob);
            try {
                w.location.href = url;
                // Add listener for window close AFTER navigation to avoid premature triggering
                w.addEventListener('beforeunload', () => {
                    console.log('[Projection] Window closing - beforeunload triggered');
                    URL.revokeObjectURL(url);
                    setIsProjectionOpen(false);
                    projectionWindowRef.current = null;
                });
                console.log('[Projection] Navigation complete, beforeunload listener added');
            } catch (e) {
                // Fallback: some environments may block navigation; try document.write as before.
                try {
                    w.document.open();
                    w.document.write(html);
                    w.document.close();
                    // Add listener in fallback path too
                    w.addEventListener('beforeunload', () => {
                        console.log('[Projection] Window closing - beforeunload triggered (fallback)');
                        setIsProjectionOpen(false);
                        projectionWindowRef.current = null;
                    });
                } catch (e2) {
                    URL.revokeObjectURL(url);
                    throw e2;
                }
            }
        } catch (err) {
            console.error('[Projection] Projection failed:', err);
            setError('Nepodařilo se otevřít projekční okno.');
            console.log('[Projection] Setting isProjectionOpen = false due to error');
            setIsProjectionOpen(false);
            projectionWindowRef.current = null;
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
                    disabled={!selectedSongs.length || isCombining || isProjectionOpen}
                    title={isProjectionOpen ? "Projekční okno je již otevřeno" : ""}
                >
                    Promítat texty
                </button>

                {isProjectionOpen && (
                    <div className={styles.projectionControls}>
                        <div style={{ marginTop: '12px', paddingTop: '12px', borderTop: '1px solid rgba(0,0,0,0.1)' }}>
                            <p style={{ fontSize: '12px', color: '#666', margin: '0 0 8px 0', fontWeight: 'bold' }}>Řízení projekce:</p>

                            {projectionSongsData.length > 0 && (
                                <div style={{ marginBottom: '12px', maxHeight: '200px', overflowY: 'auto', border: '1px solid rgba(0,0,0,0.1)', borderRadius: '4px', background: '#f9f9f9' }}>
                                    {projectionSongsData.map((song, songIdx) => {
                                        const sequence = song.verseOrder && song.verseOrder.trim()
                                            ? song.verseOrder.split(/\s+/).filter(Boolean)
                                            : song.verses.map(v => v.name);

                                        return (
                                            <div key={songIdx} style={{ borderBottom: songIdx < projectionSongsData.length - 1 ? '1px solid rgba(0,0,0,0.1)' : 'none', padding: '8px' }}>
                                                <div style={{ fontSize: '11px', fontWeight: 'bold', marginBottom: '4px', color: '#333' }}>{song.title}</div>
                                                <div style={{ display: 'flex', flexWrap: 'wrap', gap: '4px' }}>
                                                    {sequence.map((verseName, verseIdx) => {
                                                        const verseObj = song.verses.find(v => v.name === verseName) || song.verses[verseIdx];
                                                        if (!verseObj) return null;

                                                        const firstLine = (verseObj.lines || '').split('\n')[0] || '';
                                                        const isActive = songIdx === currentSongIdx && verseIdx === currentVerseIdx;

                                                        return (
                                                            <div
                                                                key={verseIdx}
                                                                style={{
                                                                    fontSize: '10px',
                                                                    padding: '4px 6px',
                                                                    background: isActive ? '#4CAF50' : '#fff',
                                                                    color: isActive ? '#fff' : '#666',
                                                                    border: '1px solid rgba(0,0,0,0.1)',
                                                                    borderRadius: '3px',
                                                                    flex: '1 1 100%',
                                                                    maxWidth: '100%',
                                                                    fontWeight: isActive ? 'bold' : 'normal'
                                                                }}
                                                            >
                                                                <strong>{verseName}:</strong> {firstLine.substring(0, 40)}{firstLine.length > 40 ? '...' : ''}
                                                            </div>
                                                        );
                                                    })}
                                                </div>
                                            </div>
                                        );
                                    })}
                                </div>
                            )}
                            <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
                                <button
                                    type="button"
                                    className={styles.actionButton}
                                    style={{ fontSize: '12px', padding: '6px 10px', flex: '1 1 calc(50% - 4px)' }}
                                    onClick={() => sendProjectionCommand('prevVerse')}
                                    title="Předchozí verš"
                                >
                                    ◀︎ Sloka
                                </button>
                                <button
                                    type="button"
                                    className={styles.actionButton}
                                    style={{ fontSize: '12px', padding: '6px 10px', flex: '1 1 calc(50% - 4px)' }}
                                    onClick={() => sendProjectionCommand('nextVerse')}
                                    title="Další verš"
                                >
                                    Sloka ▶︎
                                </button>
                                <button
                                    type="button"
                                    className={styles.actionButton}
                                    style={{ fontSize: '12px', padding: '6px 10px', flex: '1 1 calc(50% - 4px)' }}
                                    onClick={() => sendProjectionCommand('prevSong')}
                                    title="Předchozí píseň"
                                >
                                    ◀︎ Píseň
                                </button>
                                <button
                                    type="button"
                                    className={styles.actionButton}
                                    style={{ fontSize: '12px', padding: '6px 10px', flex: '1 1 calc(50% - 4px)' }}
                                    onClick={() => sendProjectionCommand('nextSong')}
                                    title="Další píseň"
                                >
                                    Píseň ▶︎
                                </button>
                                <button
                                    type="button"
                                    className={styles.actionButton}
                                    style={{ fontSize: '12px', padding: '6px 10px', flex: '1 1 100%', background: '#f44336', color: '#fff' }}
                                    onClick={closeProjection}
                                    title="Zavřít projekční okno"
                                >
                                    ✕ Zavřít projektor
                                </button>
                            </div>
                        </div>
                    </div>
                )}

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
