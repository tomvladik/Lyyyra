import { useContext, useEffect, useState } from 'react';
import * as go from '../../../wailsjs/go/main/App';

import { SongCard } from "../../components/SongCard";
import { DataContext } from '../../context';
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
    
    useEffect(() => {
        // Poll for new songs while database is being filled
        if (status.IsProgress && status.DatabaseReady === false && status.SongsReady === true) {
            const pollInterval = setInterval(() => {
                fetchData();
            }, 7777);
            return () => clearInterval(pollInterval);
        }
    }, [status.IsProgress, status.DatabaseReady, status.SongsReady]);
    
    useEffect(() => {
        // Delay action after page render (500ms delay in this case)
        const timer = setTimeout(() => {
            console.log('This runs after 500ms delay');
            fetchData()
        }, 500);

        // Cleanup function to clear the timeout if the component unmounts
        return () => clearTimeout(timer);
    }, []);

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
