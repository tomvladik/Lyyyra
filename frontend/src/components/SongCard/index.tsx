import { useEffect, useState } from "react";
import { GetSongAuthors } from "../../../wailsjs/go/main/App";
import { Author, dtoSong } from "../../models";
import HighlightText from "../HighlightText";
import styles from "./index.module.less";

export const SongCard = ({ data }: { data: dtoSong }) => {
    const [authorData, setData] = useState(new Array<Author>());
    const [, setError] = useState(false);

    const fetchData = async () => {
        try {
            // Assume fetchData returns a Promise
            const result = await GetSongAuthors(data.Id);
            setData(result);
        } catch (error) {
            setError(true);
        }
    };
    // useEffect with an empty dependency array runs once when the component mounts
    useEffect(() => {
        fetchData();
    }, []); // Empty dependency array means it runs once when the component mounts

    return (
        <div className={styles.songCard}>
            <div className={styles.songHeader}>
                <div className={styles.title}><span className={styles.songNumber}>{data.Entry}:</span> {data.Title}</div>
                {authorData?.filter((el) => el.Type === "words")
                    .map((auth) => {
                        return <div key={"T-" + auth.Value} className={styles.author}><b>T:</b> {auth.Value}</div>;
                    })}
                {authorData?.filter((el) => el.Type === "music")
                    .map((auth) => {
                        return <div key={"M-" + auth.Value} className={styles.author}><b>M:</b> {auth.Value}</div>;
                    })}

            </div>
            <div className={styles.lyrics2} style={{ marginBottom: '1px' }}>
                {data.Verses.split('\n').map((paragraph, index) => (
                    <HighlightText key={index} text={paragraph} />
                ))}
            </div>
        </div>
    );
};