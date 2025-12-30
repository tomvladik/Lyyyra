import { useContext, useEffect, useState } from 'react';
import * as go from '../../../wailsjs/go/main/App';

import { SongCard } from "../../components/SongCard";
import { INITIAL_LOAD_DELAY, SONG_POLL_INTERVAL } from '../../constants';
import { DataContext } from '../../context';
import { useDelayedEffect, usePolling } from '../../hooks/usePolling';
import { dtoSong } from '../../models';
import { removeDiacritics } from "../../utils/stringUtils";


export const SongList = () => {
    const { status } = useContext(DataContext);
    const [songs, setSongs] = useState(new Array<dtoSong>());

    const fetchData = async () => {
        try {
            // Assume fetchData returns a Promise
            const songs = await go.GetSongs(status.Sorting, status.SearchPattern);
            setSongs(songs);
        } catch (error) {
            console.log(error)
        }
    };
    useEffect(() => {
        fetchData()
    }, [status.SearchPattern, status.Sorting]);
    
    // Poll for new songs while database is being filled
    const shouldPoll = status.IsProgress && !status.DatabaseReady && status.SongsReady;
    usePolling(fetchData, SONG_POLL_INTERVAL, shouldPoll);
    
    // Delay action after page render
    useDelayedEffect(() => {
        console.log('Initial song list load');
        fetchData();
    }, INITIAL_LOAD_DELAY, []);

    const normalizedFilter = removeDiacritics(status.SearchPattern)?.toLowerCase();
    return (
        <div>
            {songs
                ?.filter((el) => {
                    if (normalizedFilter.length < 3) return true;
                    return removeDiacritics(el.Title).toLowerCase().includes(normalizedFilter)
                        || removeDiacritics(el.Verses).toLowerCase().includes(normalizedFilter)
                        || removeDiacritics(el.AuthorMusic).toLowerCase().includes(normalizedFilter)
                        || removeDiacritics(el.AuthorLyric).toLowerCase().includes(normalizedFilter)
                        ;

    ;
                })
                .map((song) => {
                    return <SongCard key={song.Id} data={song} />;
                })}
        </div>
    );
};
