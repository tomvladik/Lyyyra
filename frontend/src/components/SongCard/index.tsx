import { useContext, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { GetSongAuthors } from "../../../wailsjs/go/app/App";
import { Author, dtoSong } from "../../models";
import { SelectionContext } from "../../selectionContext";
import { parseVerses } from "../../utils/verseUtils";
import HighlightText from "../HighlightText";
import { PdfModal } from "../PdfModal";
import { AuthorList } from "./AuthorList";
import styles from "./index.module.less";

export const SongCard = ({ data }: { data: dtoSong }) => {
    const { t } = useTranslation();
    const [authorData, setData] = useState(new Array<Author>());
    const [pdfModalOpen, setPdfModalOpen] = useState(false);
    const { addSongToSelection, isSongSelected } = useContext(SelectionContext);

    const fetchData = async () => {
        try {
            // Assume fetchData returns a Promise
            const result = await GetSongAuthors(data.Id);
            setData(result);
        } catch (error) {
            console.error("Failed to fetch song authors:", error);
        }
    };

    const handleOpenPdf = () => {
        if (data.KytaraFile) {
            setPdfModalOpen(true);
        }
    };

    const isSelected = isSongSelected(data.Id);

    const handleAddToSelection = () => {
        if (isSelected) {
            return;
        }

        addSongToSelection({
            id: data.Id,
            entry: data.Entry,
            title: data.Title,
            filename: data.KytaraFile || undefined,
            hasNotes: !!data.KytaraFile,
        });
    };

    // useEffect with an empty dependency array runs once when the component mounts
    useEffect(() => {
        fetchData();
    }, []); // fetchData recreated on every render, but we only want to call it once on mount

    return (
        <>
            <div className={styles.songCard}>
                <div className={styles.songHeader}>
                    <div className={styles.title}>
                        <span className={styles.songNumber}>{data.Entry}:</span>{' '}
                        <HighlightText as="span" text={data.Title} />
                    </div>
                    <AuthorList authors={authorData} type="words" />
                    <AuthorList authors={authorData} type="music" />
                </div>
                <div className={styles.lyrics}>
                    {parseVerses(data.Verses).map((verse, idx) => (
                        <div key={idx}>
                            <HighlightText text={verse} />
                        </div>
                    ))}
                </div>
                <div className={styles.songFooterRow}>
                    <div className={styles.songFooter}>
                        {data.KytaraFile && (
                            <span
                                className={styles.actionIcon}
                                onClick={handleOpenPdf}
                                title={t('songCard.showNotes')}
                            >
                                🎵
                            </span>
                        )}
                        {!data.KytaraFile && (
                            <span
                                className={[styles.actionIcon, styles.actionIconDisabled].join(" ")}
                                title={t('songCard.notesUnavailable')}
                                aria-disabled="true"
                            >
                                🎵
                            </span>
                        )}
                        <span
                            className={[styles.actionIcon, isSelected ? styles.actionIconDisabled : ""].join(" ").trim()}
                            onClick={handleAddToSelection}
                            title={isSelected ? t('songCard.alreadyInSelection') : t('songCard.addToSelection')}
                            aria-disabled={isSelected}
                        >
                            📋
                        </span>
                    </div>
                </div>
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