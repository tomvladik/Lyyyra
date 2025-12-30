import { FC, useContext } from 'react';
import { DataContext } from '../../context';
import { removeDiacritics } from '../../utils/stringUtils';

interface HighlightTextProps {
   text: string;
}

const HighlightText: FC<HighlightTextProps> = ({ text }) => {
   const { status } = useContext(DataContext);

   if (status.SearchPattern === '') {
      return <p>{text}</p>;
   }

   const normalizedPattern = removeDiacritics(status.SearchPattern).toLowerCase();
   const normalizedText = removeDiacritics(text).toLowerCase();
   const regex = new RegExp(`(${normalizedPattern})`, 'gi');

   // Split the normalized text into an array, where matches are separated
   const parts = normalizedText.split(regex);

   // Map the parts back to the original text
   let originalIndex = 0;
   const originalParts = parts.map((part) => {
      const originalPart = text.slice(originalIndex, originalIndex + part.length);
      originalIndex += part.length;
      return originalPart;
   });

   return (
      <p>
         {originalParts.map((part: string, index: number) =>
            regex.test(removeDiacritics(part)) ? <mark key={index}>{part}</mark> : part
         )}
      </p>
   );
};

export default HighlightText;
