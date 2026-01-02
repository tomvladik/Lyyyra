import { useContext, useEffect, useState } from "react";
import { GetSongAuthors } from "../../../wailsjs/go/main/App";
import { Author, dtoSong, SongPdfVariant } from "../../models";
import { SelectionContext } from "../../selectionContext";
import HighlightText from "../HighlightText";
import { PdfModal } from "../PdfModal";
import styles from "./index.module.less";

export const SongCard = ({ data }: { data: dtoSong }) => {
    const [authorData, setData] = useState(new Array<Author>());
    const [, setError] = useState(false);
    const [pdfModalVariant, setPdfModalVariant] = useState<SongPdfVariant | null>(null);
    const { addSongToSelection, getSelectedSong } = useContext(SelectionContext);

    const fetchData = async () => {
        try {
            // Assume fetchData returns a Promise
            const result = await GetSongAuthors(data.Id);
            setData(result);
        } catch (error) {
            setError(true);
        }
    };

    const getFilenameForVariant = (variant: SongPdfVariant) =>
        variant === "guitar" ? data.KytaraFile : data.NotesFile;

    const handleOpenPdf = (variant: SongPdfVariant) => {
        const file = getFilenameForVariant(variant);
        if (!file) {
            return;
        }
        setPdfModalVariant(variant);
    };

    const selectedSong = getSelectedSong(data.Id);
    const selectedVariant = selectedSong?.variant;

    const handleAddToSelection = (variant: SongPdfVariant) => {
        const filename = getFilenameForVariant(variant);
        if (!filename) {
            return;
        }

        addSongToSelection({
            id: data.Id,
            entry: data.Entry,
            title: data.Title,
            filename,
            variant,
        });
    };

    // useEffect with an empty dependency array runs once when the component mounts
    useEffect(() => {
        fetchData();
    }, []); // Empty dependency array means it runs once when the component mounts

    return (
        <>
            <div className={styles.songCard}>
                <div className={styles.songHeader}>
                    <div className={styles.title}>
                        <span className={styles.songNumber}>{data.Entry}:</span>{' '}
                        <HighlightText as="span" text={data.Title} />
                    </div>
                    {authorData?.filter((el) => el.Type === "words")
                        .map((auth) => {
                            return (
                                <div key={"T-" + auth.Value} className={styles.author}>
                                    <b>T:</b> <HighlightText as="span" text={auth.Value} />
                                </div>
                            );
                        })}
                    {authorData?.filter((el) => el.Type === "music")
                        .map((auth) => {
                            return (
                                <div key={"M-" + auth.Value} className={styles.author}>
                                    <b>M:</b> <HighlightText as="span" text={auth.Value} />
                                </div>
                            );
                        })}

                </div>
                <div className={styles.lyrics} style={{ marginBottom: '1px' }}>
                    {(() => {
                        const verses = (data.Verses || '')
                            .split(/\n\n+/)
                            .map(v => v.replace(/\r?\n+/g, ' ').replace(/\s+/g, ' ').trim())
                            .filter(Boolean);
                        return verses.map((verse, idx) => (
                            <div key={idx}>
                                <HighlightText text={verse} />
                            </div>
                        ));
                    })()}
                </div>
                {(data.KytaraFile || data.NotesFile) && (
                    <div className={styles.songFooterRow}>
                        <div className={styles.songFooter}>
                            {data.KytaraFile && (
                                <>
                                    <span
                                        className={styles.actionIcon}
                                        onClick={() => handleOpenPdf("guitar")}
                                        title="Zobrazit kytarové noty"
                                    >
                                        🎸
                                    </span>
                                    <span
                                        className={[styles.actionIcon, selectedVariant === "guitar" ? styles.actionIconDisabled : ""].join(" ").trim()}
                                        onClick={() => handleAddToSelection("guitar")}
                                        title={selectedVariant === "guitar" ? "Kytarové noty už jsou ve výběru" : "Přidat kytarové noty do výběru"}
                                        aria-disabled={selectedVariant === "guitar"}
                                    >
                                        📋
                                    </span>
                                </>
                            )}
                            {data.NotesFile && (
                                <>
                                    <span
                                        className={styles.actionIcon}
                                        onClick={() => handleOpenPdf("notes")}
                                        title="Zobrazit noty (bez kytary)"
                                    >
                                        🎼
                                    </span>
                                    <span
                                        className={[styles.actionIcon, selectedVariant === "notes" ? styles.actionIconDisabled : ""].join(" ").trim()}
                                        onClick={() => handleAddToSelection("notes")}
                                        title={selectedVariant === "notes" ? "Tyto noty už jsou ve výběru" : "Přidat noty (bez kytary) do výběru"}
                                        aria-disabled={selectedVariant === "notes"}
                                    >
                                        📄
                                    </span>
                                </>
                            )}
                        </div>
                    </div>
                )}
            </div>
            <PdfModal
                isOpen={!!pdfModalVariant}
                filename={pdfModalVariant === "guitar" ? data.KytaraFile || undefined : data.NotesFile || undefined}
                songNumber={data.Entry}
                songName={data.Title}
                onClose={() => setPdfModalVariant(null)}
            />
        </>
    );
};