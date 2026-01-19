import { useContext, useEffect, useState } from 'react';
import * as go from '../../../wailsjs/go/app/App';

import { SongCard } from "../../components/SongCard";
import { INITIAL_LOAD_DELAY, SONG_POLL_INTERVAL } from '../../constants';
import { DataContext } from '../../context';
import { useDelayedEffect, usePolling } from '../../hooks/usePolling';
import { dtoSong } from '../../models';


export const SongList = () => {
    const { status } = useContext(DataContext);
    const [songs, setSongs] = useState(new Array<dtoSong>());

    const fetchData = async () => {
        try {
            // Assume fetchData returns a Promise
            const songs = await go.GetSongs(status.Sorting, status.SearchPattern);
            setSongs(songs);
        } catch (error) {
            console.error("Failed to fetch songs:", error);
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
        fetchData();
    }, INITIAL_LOAD_DELAY, []);

    return (
        <div>
            {songs
                ?.map((song) => {
                    return <SongCard key={song.Id} data={song} />;
                })}
        </div>
    );
};
