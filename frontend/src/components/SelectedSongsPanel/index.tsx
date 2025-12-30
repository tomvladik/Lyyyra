import { useContext, useMemo, useState } from "react";
import { GetCombinedPdf } from "../../../wailsjs/go/main/App";
import { PdfModal } from "../PdfModal";
import { SelectionContext } from "../../selectionContext";
import styles from "./index.module.less";

export const SelectedSongsPanel = () => {
    const { selectedSongs, removeSongFromSelection, clearSelection } = useContext(SelectionContext);
    const [isCombining, setIsCombining] = useState(false);
    const [combinedPdf, setCombinedPdf] = useState("");
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [error, setError] = useState("");

    const panelTitle = useMemo(() => {
        if (!selectedSongs.length) {
            return "Výběr je prázdný";
        }
        if (selectedSongs.length === 1) {
            return "1 skladba ve výběru";
        }
        return `${selectedSongs.length} skladeb ve výběru`;
    }, [selectedSongs.length]);

    const handleRemove = (id: number) => {
        removeSongFromSelection(id);
    };

    const handleCombineClick = async () => {
        if (!selectedSongs.length) {
            return;
        }

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

    const handleCloseModal = () => {
        setIsModalOpen(false);
        setCombinedPdf("");
    };

    return (
        <aside className={styles.panel} aria-label="Výběr skladeb">
            <header className={styles.panelHeader}>
                <div>
                    <p className={styles.panelLabel}>Připravené noty</p>
                    <h2 className={styles.panelTitle}>{panelTitle}</h2>
                </div>
                <button
                    type="button"
                    className={styles.clearButton}
                    onClick={clearSelection}
                    disabled={!selectedSongs.length || isCombining}
                >
                    Vyčistit
                </button>
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
                    className={styles.combineButton}
                    onClick={handleCombineClick}
                    disabled={!selectedSongs.length || isCombining}
                >
                    {isCombining ? "Vytvářím PDF…" : "Zobrazit společné PDF"}
                </button>
                {error && <p className={styles.errorText}>{error}</p>}
            </div>

            <PdfModal
                isOpen={isModalOpen}
                dataUrl={combinedPdf}
                songName="Spojené noty"
                onClose={handleCloseModal}
            />
        </aside>
    );
};
