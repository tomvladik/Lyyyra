import { SongCard } from "../../components/SongCard";
import { dtoSong } from '../../models';

export const SongList = ({ songs, filter }: { songs: Array<dtoSong>, filter: string }) => {

    return (
        <div>
            {songs
                ?.filter((el) => el.Title.toLowerCase().includes(filter?.toLowerCase())
                    || el.Verses.toLowerCase().includes(filter?.toLowerCase()))
                .map((song) => {
                    return <SongCard key={song.Id} data={song} />;
                })}
        </div>
    );
};
