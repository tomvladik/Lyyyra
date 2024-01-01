import { SongData, dtoSong } from "../../models";
import styles from "./index.module.less";
export const SongCard = ({ data }: { data: dtoSong }) => {
    return (
        <div className={styles.songCard}>
            <div className={styles.title}><span className={styles.songNumber}>{data.Entry}:</span> {data.Title}</div>
            <div className={styles.author}><b>T:</b> xxx</div>
            <div className={styles.author}><b>M:</b> yyy</div>
            <div className={styles.lyrics2} style={{ marginBottom: '16px' }}>
                {data.Verses.split('\n').map((paragraph, index) => (
                    <p key={index}>{paragraph}</p>
                ))}
            </div>
        </div>
    );
};