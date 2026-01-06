import { Author } from "../../models";
import HighlightText from "../HighlightText";
import styles from "./index.module.less";

interface AuthorListProps {
    authors: Author[];
    type: "words" | "music";
}

/**
 * Renders a list of authors for a specific type (words or music).
 * Displays with appropriate label (T: for text/words, M: for music).
 */
export const AuthorList = ({ authors, type }: AuthorListProps) => {
    const label = type === "words" ? "T:" : "M:";

    if (!authors || authors.length === 0) return null;

    const filteredAuthors = authors.filter((el) => el.Type === type);

    if (filteredAuthors.length === 0) return null;

    return (
        <>
            {filteredAuthors.map((auth) => (
                <div key={`${type}-${auth.Value}`} className={styles.author}>
                    <b>{label}</b> <HighlightText as="span" text={auth.Value} />
                </div>
            ))}
        </>
    );
};
