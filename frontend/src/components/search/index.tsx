import styles from "./index.module.less";

export const Search = ({ onChange }: { onChange: React.ChangeEventHandler }) => {
    return <input id="search-box" className={styles.search} type="text" onChange={onChange} placeholder="Hledat text ..." />;
};