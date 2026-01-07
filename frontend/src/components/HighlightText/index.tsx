import { FC, useContext, useMemo } from 'react';
import { DataContext } from '../../context';
import { removeDiacritics } from '../../utils/stringUtils';

type WrapperTag = 'p' | 'span' | 'div';

interface HighlightTextProps {
   text: string;
   as?: WrapperTag;
   className?: string;
}

const HighlightText: FC<HighlightTextProps> = ({ text, as = 'p', className }) => {
   const { status } = useContext(DataContext);
   const normalizedPattern = useMemo(() => removeDiacritics(status.SearchPattern?.normalize('NFC') || '').toLowerCase(), [status.SearchPattern]);

   const Wrapper = as as React.ElementType;

   if (!normalizedPattern) {
      return <Wrapper className={className}>{text}</Wrapper>;
   }

   const normalizedText = removeDiacritics(text.normalize('NFC')).toLowerCase();
   const escapedPattern = normalizedPattern.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
   const regex = new RegExp(`(${escapedPattern})`, 'gi');

   // Create a mapping from normalized position to original position
   const normalizedToOriginal: number[] = [];
   for (let i = 0; i < text.length; i++) {
      const origChar = text[i].normalize('NFC');
      const normalized = removeDiacritics(origChar).toLowerCase();
      for (let j = 0; j < normalized.length; j++) {
         normalizedToOriginal.push(i);
      }
   }

   const segments = normalizedText.split(regex);

   let normCursor = 0;
   const mappedSegments = segments.map((segment) => {
      const segmentLength = segment.length;
      const origStart = normalizedToOriginal[normCursor] ?? 0;
      const origEnd = normCursor + segmentLength < normalizedToOriginal.length
         ? normalizedToOriginal[normCursor + segmentLength - 1] + 1
         : text.length;
      const originalSegment = text.slice(origStart, origEnd);
      normCursor += segmentLength;
      const isMatch = segment !== '' && segment === normalizedPattern;
      return { originalSegment, isMatch };
   });

   return (
      <Wrapper className={className}>
         {mappedSegments.map((segment, index) =>
            segment.isMatch ? (
               <mark key={`${segment.originalSegment}-${index}`}>{segment.originalSegment}</mark>
            ) : (
               <span key={`${segment.originalSegment}-${index}`}>{segment.originalSegment}</span>
            )
         )}
      </Wrapper>
   );
};

export default HighlightText;
