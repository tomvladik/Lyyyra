import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import '@testing-library/jest-dom/vitest';
import HighlightText from '../index';
import { DataContext } from '../../../context';

describe('<HighlightText />', () => {
  const renderWithContext = (text: string, searchPattern: string, as?: 'p' | 'span' | 'div') => {
    const mockStatus = {
      DatabaseReady: true,
      SongsReady: true,
      WebResourcesReady: true,
      IsProgress: false,
      LastSave: '',
      SearchPattern: searchPattern,
      Sorting: 'entry' as const,
      ProgressMessage: '',
      ProgressPercent: 0,
    };
    const mockUpdateStatus = vi.fn();

    return render(
      <DataContext.Provider value={{ status: mockStatus, updateStatus: mockUpdateStatus }}>
        <HighlightText text={text} as={as} />
      </DataContext.Provider>
    );
  };

  it('renders plain text when search pattern is empty', () => {
    renderWithContext('Hello world', '');
    const paragraph = screen.getByText('Hello world');
    expect(paragraph.tagName).toBe('P');
    expect(paragraph.querySelector('mark')).toBeNull();
  });

  it('highlights matching text', () => {
    renderWithContext('Hello world', 'world');
    const mark = screen.getByText('world');
    expect(mark.tagName).toBe('MARK');
  });

  it('highlights text case-insensitively', () => {
    renderWithContext('Hello World', 'world');
    const mark = screen.getByText('World');
    expect(mark.tagName).toBe('MARK');
  });

  it('handles diacritics in search pattern', () => {
    const { container } = renderWithContext('Zpěvník', 'zpevnik');
    const mark = container.querySelector('mark');
    expect(mark).toBeInTheDocument();
    expect(mark?.textContent).toBe('Zpěvník');
  });

  it('handles diacritics in text', () => {
    const { container } = renderWithContext('Zpěvník', 'Zpěvník');
    const mark = container.querySelector('mark');
    expect(mark).toBeInTheDocument();
    expect(mark?.textContent).toBe('Zpěvník');
  });

  it('highlights multiple occurrences', () => {
    const { container } = renderWithContext('test test test', 'test');
    const marks = container.querySelectorAll('mark');
    expect(marks).toHaveLength(3);
  });

  it('preserves original text case in highlights', () => {
    renderWithContext('Hello HELLO hello', 'hello');
    expect(screen.getByText('Hello')).toBeInTheDocument();
    expect(screen.getByText('HELLO')).toBeInTheDocument();
    expect(screen.getByText('hello')).toBeInTheDocument();
  });

  it('does not highlight when pattern does not match', () => {
    const { container } = renderWithContext('Hello world', 'xyz');
    const mark = container.querySelector('mark');
    expect(mark).toBeNull();
  });

  it('supports custom wrapper tags', () => {
    renderWithContext('Inline text', '', 'span');
    const element = screen.getByText('Inline text');
    expect(element.tagName).toBe('SPAN');
  });
});
