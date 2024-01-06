import { SongCard } from "../../components/SongCard";
import { dtoSong } from '../../models';

export const SongList = ({ songs, inputValue }: { songs: Array<dtoSong>, inputValue: string }) => {

    return (
        <div>


            {songs
                ?.filter((el) => el.Title.toLowerCase().includes(inputValue.toLowerCase())
                    || el.Verses.toLowerCase().includes(inputValue.toLowerCase()))
                .map((song) => {
                    return <SongCard data={song} />;
                })}
        </div>
    );
};
