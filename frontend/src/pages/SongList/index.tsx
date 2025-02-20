import { SongCard } from "../../components/SongCard";
import { dtoSong } from '../../models';
import { removeDiacritics } from "../../utils/stringUtils";


export const SongList = ({ songs, filter }: { songs: Array<dtoSong>, filter: string }) => {
    const normalizedFilter = removeDiacritics(filter)?.toLowerCase();
    return (
        <div>
            {songs
                ?.filter((el) => {
                    if (normalizedFilter.length < 3) return true;
                    return removeDiacritics(el.Title).toLowerCase().includes(normalizedFilter)
                        || removeDiacritics(el.Verses).toLowerCase().includes(normalizedFilter);
                })
                .map((song) => {
                    return <SongCard key={song.Id} data={song} />;
                })}
        </div>
    );
};
