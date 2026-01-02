import { useContext, useEffect, useState } from "react";
import { GetSongAuthors } from "../../../wailsjs/go/main/App";
import { Author, dtoSong } from "../../models";
import { SelectionContext } from "../../selectionContext";
import HighlightText from "../HighlightText";
import { PdfModal } from "../PdfModal";
import styles from "./index.module.less";

export const SongCard = ({ data }: { data: dtoSong }) => {
    const [authorData, setData] = useState(new Array<Author>());
    const [, setError] = useState(false);
    const [pdfModalOpen, setPdfModalOpen] = useState(false);
    const { addSongToSelection, isSongSelected } = useContext(SelectionContext);

    const fetchData = async () => {
        try {
            // Assume fetchData returns a Promise
            const result = await GetSongAuthors(data.Id);
            setData(result);
        } catch (error) {
            setError(true);
        }
    };

    const handleOpenPdf = () => {
        if (data.KytaraFile) {
            setPdfModalOpen(true);
        }
    };

    const isSelected = isSongSelected(data.Id);

    const handleAddToSelection = () => {
        if (!data.KytaraFile || isSelected) {
            return;
        }

        addSongToSelection({
            id: data.Id,
            entry: data.Entry,
            title: data.Title,
            filename: data.KytaraFile,
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
                {data.KytaraFile && (
                    <div className={styles.songFooterRow}>
                        <div className={styles.songFooter}>
                            <span
                                className={styles.actionIcon}
                                onClick={handleOpenPdf}
                                title="Zobrazit noty"
                            >
                                🎵
                            </span>
                            <span
                                className={[styles.actionIcon, isSelected ? styles.actionIconDisabled : ""].join(" ").trim()}
                                onClick={handleAddToSelection}
                                title={isSelected ? "Skladba už je ve výběru" : "Přidat do výběru"}
                                aria-disabled={isSelected}
                            >
                                📋
                            </span>
                        </div>
                    </div>
                )}
            </div>
            <PdfModal
                isOpen={pdfModalOpen}
                filename={data.KytaraFile || undefined}
                songNumber={data.Entry}
                songName={data.Title}
                onClose={() => setPdfModalOpen(false)}
            />
        </>
    );
};