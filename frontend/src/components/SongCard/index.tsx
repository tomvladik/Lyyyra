import { useEffect, useState } from "react";
import { GetSongAuthors } from "../../../wailsjs/go/main/App";
import { Author, dtoSong } from "../../models";
import HighlightText from "../HighlightText";
import { PdfModal } from "../PdfModal";
import styles from "./index.module.less";

export const SongCard = ({ data }: { data: dtoSong }) => {
    const [authorData, setData] = useState(new Array<Author>());
    const [, setError] = useState(false);
    const [pdfModalOpen, setPdfModalOpen] = useState(false);

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

    const handleSecondAction = () => {
        // Placeholder for second action
        console.log("Second action for:", data.KytaraFile);
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
            <div className={styles.lyrics2} style={{ marginBottom: '1px' }}>
                {data.Verses.split('\n').map((paragraph, index) => (
                    <HighlightText key={index} text={paragraph} />
                ))}
            </div>
            {data.KytaraFile && (
                <div className={styles.songFooter}>
                    <span 
                        className={styles.actionIcon} 
                        onClick={handleOpenPdf}
                        title="Otevřít PDF"
                    >
                        🎵
                    </span>
                    <span 
                        className={styles.actionIcon}
                        onClick={handleSecondAction}
                        title="Další akce"
                    >
                        📋
                    </span>
                </div>
            )}
        </div>
            <PdfModal 
                isOpen={pdfModalOpen} 
                filename={data.KytaraFile || ""} 
                onClose={() => setPdfModalOpen(false)} 
            />
        </>
    );
};