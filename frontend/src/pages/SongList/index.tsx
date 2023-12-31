import { useState } from "react";

import { Filter } from "../../components/filter";
import { Search } from "../../components/search";
import { SongCard } from "../../components/SongCard";

import styles from "./index.module.less";
import { loremIpsum } from "lorem-ipsum";
import { SongData } from "../../models";



export const SongList = () => {
    const [inputValue, setInputValue] = useState("");
    const [activeFilter, setActiveFilter] = useState("");

    //   const posts = useMany<{
    //     id: number;
    //     title: string;
    //     status: string;
    //   }>({
    //     resource: "posts",
    //     ids: Array.from(Array(8).keys()).slice(1),
    //   }).data?.data;


    function generateItems(): SongData[] {
        const items: SongData[] = [];

        for (let i = 1; i <= 30; i++) {
            const item: SongData = {
                songNumber: i,
                title: loremIpsum({
                    count: 1,                // Number of "words", "sentences", or "paragraphs"
                    sentenceLowerBound: 3,   // Min. number of words per sentence.
                    sentenceUpperBound: 9,  // Max. number of words per sentence."\r\n" (win32)
                    units: "sentences",      // paragraph(s), "sentence(s)", or "word(s)"
                }),
                authorText: loremIpsum({
                    count: 1,                // Number of "words", "sentences", or "paragraphs"
                    sentenceLowerBound: 2,   // Min. number of words per sentence.
                    sentenceUpperBound: 4,  // Max. number of words per sentence."\r\n" (win32)
                    units: "sentences",      // paragraph(s), "sentence(s)", or "word(s)"
                }),
                authorMusic: loremIpsum({
                    count: 1,                // Number of "words", "sentences", or "paragraphs"
                    sentenceLowerBound: 2,   // Min. number of words per sentence.
                    sentenceUpperBound: 4,  // Max. number of words per sentence."\r\n" (win32)
                    units: "sentences",      // paragraph(s), "sentence(s)", or "word(s)"
                }),
                lyrics: loremIpsum({
                    count: 8,                // Number of "words", "sentences", or "paragraphs"
                    sentenceLowerBound: 2,   // Min. number of words per sentence.
                    sentenceUpperBound: 4,  // Max. number of words per sentence."\r\n" (win32)
                    units: "sentences",      // paragraph(s), "sentence(s)", or "word(s)"
                }),
            };

            items.push(item);
        }

        return items;
    }

    const posts: SongData[] = generateItems();
    console.log(posts);

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
            {posts
                ?.filter((el) => el.title.toLowerCase().includes(inputValue.toLowerCase()))
                .map((song) => {
                    return <SongCard data={song} title={song.title} music={song.authorMusic} text={song.authorText} lyrics={song.lyrics} />;
                })}
        </div>
    );
};
