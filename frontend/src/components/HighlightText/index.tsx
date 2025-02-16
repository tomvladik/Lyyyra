import { FC, useContext } from 'react';
import { DataContext } from '../../main';

interface HighlightTextProps {
   text: string;
}

const HighlightText: FC<HighlightTextProps> = ({ text }) => {
   const { status } = useContext(DataContext);

   if (status.SearchPattern === '') {
      return <p>{text}</p>;
   }
   const regex = new RegExp(`(${status.SearchPattern})`, 'gi');

   // Split the text into an array, where matches are separated
   const parts = text.split(regex);

   return (
      <p>
         {parts.map((part: string, index: number) => regex.test(part) ? <mark key={index}>{part}</mark> : part
         )}
      </p>
   );
};

export default HighlightText;
