import { SongData } from "../../models";
import styles from "./index.module.less";
export const SongCard = ({ data, title, text, music, lyrics }: { data: SongData, title: string, text: string, music: string, lyrics: string }) => {
    return (
        <div className={styles.songCard}>
            <div className={styles.title}><span className={styles.songNumber}>{data.songNumber}:</span> {title}</div>
            <div className={styles.author}><b>T:</b> {text}</div>
            <div className={styles.author}><b>M:</b> {music}</div>
            <div className={styles.lyrics2}>{lyrics}</div>
        </div>
    );
};