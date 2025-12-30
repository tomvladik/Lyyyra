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
   const normalizedPattern = useMemo(() => removeDiacritics(status.SearchPattern || '').toLowerCase(), [status.SearchPattern]);

   const Wrapper = as as keyof JSX.IntrinsicElements;

   if (!normalizedPattern) {
      return <Wrapper className={className}>{text}</Wrapper>;
   }

   const normalizedText = removeDiacritics(text).toLowerCase();
   const escapedPattern = normalizedPattern.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
   const regex = new RegExp(`(${escapedPattern})`, 'gi');
   const segments = normalizedText.split(regex);

   let cursor = 0;
   const mappedSegments = segments.map((segment) => {
      const originalSegment = text.slice(cursor, cursor + segment.length);
      cursor += segment.length;
      const isMatch = segment !== '' && removeDiacritics(originalSegment).toLowerCase() === normalizedPattern;
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
