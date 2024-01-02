import React, { useContext, useState } from 'react';

import { Search } from "../../components/search";
import { SongCard } from "../../components/SongCard";

import { DataContext } from '../../App';

export const SongList = () => {

    const [inputValue, setInputValue] = useState("");
    const data = useContext(DataContext);


    const filters: string[] = ["published", "draft", "rejected"];

    return (
        <div>
            {/* <div className={styles.filters}>
                {filters.map((filter, index) => {
                    return (
                        <Filter
                            key={index}
                            title={filter}
                            isActive={filter === activeFilter}
                            onClick={(e: React.MouseEvent) => {
                                const el = e.target as HTMLElement;
                                el.textContent?.toLowerCase() !== activeFilter ? setActiveFilter(filter) : setActiveFilter("");
                            }}
                        />
                    );
                })}
            </div> */}
            <Search
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                    setInputValue(e.target.value);
                }}
            />
            {data.songs
                ?.filter((el) => el.Title.toLowerCase().includes(inputValue.toLowerCase())
                    || el.Verses.toLowerCase().includes(inputValue.toLowerCase()))
                .map((song) => {
                    return <SongCard data={song} />;
                })}
        </div>
    );
};
