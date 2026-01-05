import { useContext, useEffect, useMemo, useRef, useState } from "react";
import { GetCombinedPdf, GetSongProjection, GetSongVerses } from "../../../wailsjs/go/app/App";
import logoImage from "../../assets/images/logo-universal.png";
import { SelectionContext } from "../../selectionContext";
import { PdfModal } from "../PdfModal";
import styles from "./index.module.less";
import projectionTemplate from "./projection-template.html?raw";

export const SelectedSongsPanel = () => {
    const { selectedSongs, removeSongFromSelection, clearSelection } = useContext(SelectionContext);
    const panelRef = useRef<HTMLElement | null>(null);
    const projectionWindowRef = useRef<Window | null>(null);
    const projectionListRef = useRef<HTMLDivElement | null>(null);
    const activeVerseRef = useRef<HTMLDivElement | null>(null);
    const [isCombining, setIsCombining] = useState(false);
    const [combinedPdf, setCombinedPdf] = useState("");
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [error, setError] = useState("");
    const [isProjectionOpen, setIsProjectionOpen] = useState(false);
    const [projectionSongsData, setProjectionSongsData] = useState<Array<{ title: string; verseOrder: string; verses: Array<{ name: string; lines: string }> }>>([]);
    const [currentSongIdx, setCurrentSongIdx] = useState(0);
    const [currentVerseIdx, setCurrentVerseIdx] = useState(0);
    const [availableScreens, setAvailableScreens] = useState<ScreenDetailed[]>([]);
    const [showScreenSelector, setShowScreenSelector] = useState(false);
    const [selectedScreenIndex, setSelectedScreenIndex] = useState(0);

    type ScreenDetailed = {
        left: number;
        top: number;
        width: number;
        height: number;
        isPrimary: boolean;
        label: string;
    };

    // Detect available screens on mount
    useEffect(() => {
        const detectScreens = async () => {
            try {
                // Try modern Screen Detection API
                if ('getScreenDetails' in window) {
                    const screenDetails = await (window as any).getScreenDetails();
                    const screens: ScreenDetailed[] = screenDetails.screens.map((s: any, idx: number) => ({
                        left: s.left,
                        top: s.top,
                        width: s.width,
                        height: s.height,
                        isPrimary: s.isPrimary,
                        label: s.label || `Display ${idx + 1}`
                    }));
                    setAvailableScreens(screens);
                } else {
                    // Fallback: use window.screen
                    const primaryScreen: ScreenDetailed = {
                        left: 0,
                        top: 0,
                        width: window.screen.width,
                        height: window.screen.height,
                        isPrimary: true,
                        label: 'Primary Display'
                    };
                    // Check if extended display might exist
                    if ((window.screen as any).isExtended) {
                        const secondaryScreen: ScreenDetailed = {
                            left: window.screen.width,
                            top: 0,
                            width: window.screen.width,
                            height: window.screen.height,
                            isPrimary: false,
                            label: 'Secondary Display'
                        };
                        setAvailableScreens([primaryScreen, secondaryScreen]);
                    } else {
                        setAvailableScreens([primaryScreen]);
                    }
                }
            } catch (err) {
                console.log('Screen detection not available or denied');
                // Default to single screen
                setAvailableScreens([{
                    left: 0,
                    top: 0,
                    width: window.screen.width,
                    height: window.screen.height,
                    isPrimary: true,
                    label: 'Display 1'
                }]);
            }
        };
        detectScreens();
    }, []);

    // Monitor projection window status periodically
    useEffect(() => {
        if (!isProjectionOpen) return;

        const checkInterval = setInterval(() => {
            if (projectionWindowRef.current) {
                const isClosed = projectionWindowRef.current.closed;
                if (isClosed) {
                    setIsProjectionOpen(false);
                    projectionWindowRef.current = null;
                }
            }
        }, 500);

        return () => {
            clearInterval(checkInterval);
        };
    }, [isProjectionOpen]);

    // Keep highlighted verse in view while projecting
    useEffect(() => {
        if (!isProjectionOpen) return;

        const container = projectionListRef.current;
        const active = activeVerseRef.current;
        if (!container || !active) return;

        const containerRect = container.getBoundingClientRect();
        const activeRect = active.getBoundingClientRect();
        const isAbove = activeRect.top < containerRect.top;
        const isBelow = activeRect.bottom > containerRect.bottom;

        if (isAbove || isBelow) {
            active.scrollIntoView({ block: "nearest", behavior: "smooth" });
        }
    }, [currentSongIdx, currentVerseIdx, isProjectionOpen, projectionSongsData.length]);

    const sendProjectionCommand = (command: "nextVerse" | "prevVerse" | "nextSong" | "prevSong") => {
        if (projectionWindowRef.current && !projectionWindowRef.current.closed) {
            projectionWindowRef.current.postMessage({ type: "projection-control", command }, "*");
        }
    };

    const jumpToVerse = (songIdx: number, verseIdx: number) => {
        if (projectionWindowRef.current && !projectionWindowRef.current.closed) {
            projectionWindowRef.current.postMessage({ type: "projection-jump", songIdx, verseIdx }, "*");
            setCurrentSongIdx(songIdx);
            setCurrentVerseIdx(verseIdx);
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
        const songsWithNotes = selectedSongs.filter(s => s.hasNotes);
        if (!songsWithNotes.length) {
            setError("Žádná z písní ve výběru nemá dostupné noty.");
            return;
        }
        setIsCombining(true);
        setError("");
        try {
            const filenames = songsWithNotes.map(song => song.filename).filter(Boolean) as string[];
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

        if (projectionWindowRef.current && !projectionWindowRef.current.closed) {
            setError("Projekční okno je již otevřeno.");
            return;
        }

        // Show screen selector if multiple screens available
        if (availableScreens.length > 1) {
            setShowScreenSelector(true);
            return;
        }

        // Open on default screen
        await openProjectionWindow(0);
    };

    const openProjectionWindow = async (screenIndex: number) => {
        setShowScreenSelector(false);
        const screen = availableScreens[screenIndex] || availableScreens[0];

        const w = window.open("", "_blank", `popup=yes,width=${screen.width},height=${screen.height},left=${screen.left},top=${screen.top}`);
        if (!w) {
            setError("Nelze otevřít projekční okno. Zkontrolujte blokování vyskakovacích oken.");
            return;
        }

        try {
            projectionWindowRef.current = w;
            setIsProjectionOpen(true);
            setError("");
            setCurrentSongIdx(0);
            setCurrentVerseIdx(0);

            const getProj = typeof GetSongProjection === "function" ? GetSongProjection : undefined;
            const getVerses = typeof GetSongVerses === "function" ? GetSongVerses : undefined;
            const songsData: Array<{ title: string; verseOrder: string; verses: Array<{ name: string; lines: string }> }> = [];

            for (const song of selectedSongs) {
                let payloadRaw = "";
                if (typeof getProj === "function") {
                    payloadRaw = await getProj(song.id);
                } else if (typeof getVerses === "function") {
                    const v = await getVerses(song.id);
                    const parts = (v || "").split("===").map((s: string) => s.trim()).filter(Boolean);
                    const verses = parts.map((lines: string, idx: number) => ({ name: `v${idx + 1}`, lines }));
                    songsData.push({ title: song.title, verseOrder: verses.map((_: unknown, i: number) => `v${i + 1}`).join(" "), verses });
                    continue;
                }

                if (payloadRaw) {
                    try {
                        const obj = JSON.parse(payloadRaw);
                        const verses = Array.isArray(obj.verses) ? obj.verses.map((vv: unknown) => ({ name: (vv as { name?: string }).name || "", lines: (vv as { lines?: string }).lines || "" })) : [];
                        const verseOrder = typeof obj.verse_order === "string" ? obj.verse_order : "";
                        songsData.push({ title: `${song.entry}: ${song.title}`, verseOrder, verses });
                    } catch (e) {
                        songsData.push({ title: `${song.entry}: ${song.title}`, verseOrder: "", verses: [] });
                    }
                } else {
                    songsData.push({ title: `${song.entry}: ${song.title}`, verseOrder: "", verses: [] });
                }
            }

            setProjectionSongsData(songsData);

            window.addEventListener("message", (event) => {
                if (event.data && event.data.type === "projection-state") {
                    setCurrentSongIdx(event.data.songIdx || 0);
                    setCurrentVerseIdx(event.data.verseIdx || 0);
                }
            });

            const safeSongsJson = encodeURIComponent(JSON.stringify(songsData)
                .replace(/\n/g, "\\n")
                .replace(/\r/g, "\\r")
                .replace(/\u2028/g, "\\u2028")
                .replace(/\u2029/g, "\\u2029")
                .replace(/<\/script>/g, "<\\/script>")
            );

            const html = projectionTemplate
                .replace("{{SONGS_DATA}}", safeSongsJson)
                .replace("data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==", logoImage);
            const blob = new Blob([html], { type: "text/html;charset=utf-8" });
            const url = URL.createObjectURL(blob);
            try {
                w.location.href = url;
                w.addEventListener("beforeunload", () => {
                    URL.revokeObjectURL(url);
                    setIsProjectionOpen(false);
                    projectionWindowRef.current = null;
                });
            } catch (e) {
                try {
                    w.document.open();
                    w.document.write(html);
                    w.document.close();
                    w.addEventListener("beforeunload", () => {
                        setIsProjectionOpen(false);
                        projectionWindowRef.current = null;
                    });
                } catch (e2) {
                    URL.revokeObjectURL(url);
                    throw e2;
                }
            }
        } catch (err) {
            setError("Nepodařilo se otevřít projekční okno.");
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
        <aside className={styles.panel} aria-label="Výběr skladeb" ref={panelRef}>
            <header className={styles.panelHeader}>
                <div>
                    <h1 className={styles.panelTitle}>Připravené písně</h1>
                    <p className={styles.panelLabel}>{panelTitle}</p>
                </div>
                <button
                    type="button"
                    className={styles.clearButton}
                    onClick={clearSelection}
                    disabled={!selectedSongs.length || isCombining}
                >Zrušit</button>
            </header>

            <div className={styles.panelContent}>
                {!isProjectionOpen && (
                    <div className={styles.list} role="list">
                        {selectedSongs.map(song => (
                            <div key={song.id} className={styles.listItem} role="listitem">
                                <div>
                                    <span className={styles.itemNumber}>{song.entry}.</span>
                                    <span className={styles.itemTitle}>{song.title}</span>
                                    {!song.hasNotes && (
                                        <span style={{ marginLeft: "8px", fontSize: "12px", color: "#999" }}>
                                            (bez not)
                                        </span>
                                    )}
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
                )}

                {showScreenSelector && (
                    <div style={{ padding: "12px", background: "rgba(0,0,0,0.05)", borderRadius: "8px", display: "flex", flexDirection: "column", gap: "8px" }}>
                        <p style={{ margin: 0, fontWeight: "bold", fontSize: "13px" }}>Vyberte displej pro projekci:</p>
                        {availableScreens.map((screen, idx) => (
                            <button
                                key={idx}
                                type="button"
                                className={styles.actionButton}
                                style={{ width: "100%", fontSize: "12px", height: "44px", justifyContent: "center" }}
                                onClick={() => {
                                    setSelectedScreenIndex(idx);
                                    openProjectionWindow(idx);
                                }}
                            >
                                {screen.label} {screen.isPrimary ? "(Primární)" : ""} – {screen.width}×{screen.height}
                            </button>
                        ))}
                        <button
                            type="button"
                            className={styles.clearButton}
                            style={{ width: "100%", marginTop: "4px", padding: "10px 14px", textAlign: "center" }}
                            onClick={() => setShowScreenSelector(false)}
                        >
                            Zrušit
                        </button>
                    </div>
                )}

                <div className={styles.actions}>
                    {!isProjectionOpen && !showScreenSelector && (
                        <>
                            <button
                                type="button"
                                className={styles.actionButton}
                                onClick={handleCombineClick}
                                disabled={!selectedSongs.length || !selectedSongs.some(s => s.hasNotes) || isCombining}
                                title={!selectedSongs.some(s => s.hasNotes) && selectedSongs.length ? "Vámi vybrané skladby nemají dostupné noty" : ""}
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
                        </>
                    )}

                    {isProjectionOpen && (
                    <div className={styles.projectionControls}>
                        <div style={{ marginTop: "12px", paddingTop: "12px", borderTop: "1px solid rgba(0,0,0,0.1)" }}>
                            <p style={{ fontSize: "12px", color: "#666", margin: "0 0 8px 0", fontWeight: "bold" }}>Řízení projekce:</p>

                            {projectionSongsData.length > 0 && (
                                <div
                                    ref={projectionListRef}
                                    style={{ marginBottom: "12px", maxHeight: "200px", overflowY: "auto", border: "1px solid rgba(0,0,0,0.1)", borderRadius: "4px", background: "#f9f9f9" }}
                                >
                                    {projectionSongsData.map((song, songIdx) => {
                                        const sequence = song.verseOrder && song.verseOrder.trim()
                                            ? song.verseOrder.split(/\s+/).filter(Boolean)
                                            : song.verses.map(v => v.name);

                                        return (
                                            <div key={songIdx} style={{ borderBottom: songIdx < projectionSongsData.length - 1 ? "1px solid rgba(0,0,0,0.1)" : "none", padding: "8px" }}>
                                                <div style={{ fontSize: "11px", fontWeight: "bold", marginBottom: "4px", color: "#333" }}>{song.title}</div>
                                                <div style={{ display: "flex", flexWrap: "wrap", gap: "4px" }}>
                                                    {sequence.map((verseName, verseIdx) => {
                                                        const verseObj = song.verses.find(v => v.name === verseName) || song.verses[verseIdx];
                                                        if (!verseObj) return null;

                                                        const firstLine = (verseObj.lines || "").split("\n")[0] || "";
                                                        const isActive = songIdx === currentSongIdx && verseIdx === currentVerseIdx;

                                                        return (
                                                            <div
                                                                key={verseIdx}
                                                                ref={isActive ? (el) => { activeVerseRef.current = el; } : undefined}
                                                                onClick={() => jumpToVerse(songIdx, verseIdx)}
                                                                style={{
                                                                    fontSize: "10px",
                                                                    padding: "4px 6px",
                                                                    background: isActive ? "#4CAF50" : "#fff",
                                                                    color: isActive ? "#fff" : "#666",
                                                                    border: "1px solid rgba(0,0,0,0.1)",
                                                                    borderRadius: "3px",
                                                                    flex: "1 1 100%",
                                                                    maxWidth: "100%",
                                                                    fontWeight: isActive ? "bold" : "normal",
                                                                    cursor: "pointer"
                                                                }}
                                                            >
                                                                <strong>{verseName}:</strong> {firstLine.substring(0, 40)}{firstLine.length > 40 ? "..." : ""}
                                                            </div>
                                                        );
                                                    })}
                                                </div>
                                            </div>
                                        );
                                    })}
                                </div>
                            )}
                            <div style={{ display: "flex", gap: "8px", flexWrap: "wrap" }}>
                                <button
                                    type="button"
                                    className={styles.actionButton}
                                    style={{ fontSize: "12px", padding: "6px 10px", flex: "1 1 calc(50% - 4px)" }}
                                    onClick={() => sendProjectionCommand("prevVerse")}
                                    title="Předchozí verš"
                                >
                                    ◀︎ Sloka
                                </button>
                                <button
                                    type="button"
                                    className={styles.actionButton}
                                    style={{ fontSize: "12px", padding: "6px 10px", flex: "1 1 calc(50% - 4px)" }}
                                    onClick={() => sendProjectionCommand("nextVerse")}
                                    title="Další verš"
                                >
                                    Sloka ▶︎
                                </button>
                                <button
                                    type="button"
                                    className={styles.actionButton}
                                    style={{ fontSize: "12px", padding: "6px 10px", flex: "1 1 calc(50% - 4px)" }}
                                    onClick={() => sendProjectionCommand("prevSong")}
                                    title="Předchozí píseň"
                                >
                                    ◀︎ Píseň
                                </button>
                                <button
                                    type="button"
                                    className={styles.actionButton}
                                    style={{ fontSize: "12px", padding: "6px 10px", flex: "1 1 calc(50% - 4px)" }}
                                    onClick={() => sendProjectionCommand("nextSong")}
                                    title="Další píseň"
                                >
                                    Píseň ▶︎
                                </button>
                                <button
                                    type="button"
                                    className={styles.actionButton}
                                    style={{ fontSize: "12px", padding: "6px 10px", flex: "1 1 100%", background: "#f44336", color: "#fff" }}
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
