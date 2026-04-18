import { act, fireEvent, render, screen } from '@testing-library/react';
import DebouncedSearchInput from './DebouncedSearchInput';

describe('DebouncedSearchInput', () => {
  beforeEach(() => {
    jest.useFakeTimers();
  });

  afterEach(() => {
    act(() => {
      jest.runOnlyPendingTimers();
    });
    jest.useRealTimers();
  });

  it('renders with the provided value and placeholder', () => {
    render(
      <DebouncedSearchInput
        value="hello"
        onChange={jest.fn()}
        placeholder="Search stuff..."
      />,
    );
    expect(screen.getByPlaceholderText('Search stuff...')).toHaveValue('hello');
  });

  it('does not call onChange until the debounce elapses after typing stops', () => {
    const onChange = jest.fn();
    render(<DebouncedSearchInput value="" onChange={onChange} placeholder="Search" />);

    const input = screen.getByPlaceholderText('Search');
    fireEvent.change(input, { target: { value: 'a' } });
    fireEvent.change(input, { target: { value: 'ab' } });
    fireEvent.change(input, { target: { value: 'abc' } });

    act(() => {
      jest.advanceTimersByTime(299);
    });
    expect(onChange).not.toHaveBeenCalled();

    act(() => {
      jest.advanceTimersByTime(1);
    });
    expect(onChange).toHaveBeenCalledTimes(1);
    expect(onChange).toHaveBeenLastCalledWith('abc');
  });

  it('respects a custom debounceMs', () => {
    const onChange = jest.fn();
    render(
      <DebouncedSearchInput
        value=""
        onChange={onChange}
        placeholder="Search"
        debounceMs={100}
      />,
    );

    fireEvent.change(screen.getByPlaceholderText('Search'), {
      target: { value: 'hi' },
    });

    act(() => {
      jest.advanceTimersByTime(100);
    });
    expect(onChange).toHaveBeenCalledWith('hi');
  });

  it('resets the input when the external value changes', () => {
    const onChange = jest.fn();
    const { rerender } = render(
      <DebouncedSearchInput value="initial" onChange={onChange} placeholder="Search" />,
    );

    const input = screen.getByPlaceholderText('Search') as HTMLInputElement;
    fireEvent.change(input, { target: { value: 'typed' } });
    expect(input.value).toBe('typed');

    rerender(
      <DebouncedSearchInput value="fromParent" onChange={onChange} placeholder="Search" />,
    );

    expect(input.value).toBe('fromParent');

    act(() => {
      jest.advanceTimersByTime(500);
    });
    // The in-flight debounced commit for "typed" must be dropped because the
    // committed value was overridden by the parent.
    expect(onChange).not.toHaveBeenCalled();
  });

  it('fires an immediate commit with empty string when Clear is clicked', () => {
    const onChange = jest.fn();
    render(
      <DebouncedSearchInput value="hello" onChange={onChange} placeholder="Search" />,
    );

    const clearButton = screen.getByRole('button', { name: /clear search/i });
    fireEvent.click(clearButton);

    expect(onChange).toHaveBeenCalledTimes(1);
    expect(onChange).toHaveBeenCalledWith('');
    expect(
      (screen.getByPlaceholderText('Search') as HTMLInputElement).value,
    ).toBe('');
  });

  it('hides the clear button when showClear is false', () => {
    render(
      <DebouncedSearchInput
        value="hello"
        onChange={jest.fn()}
        placeholder="Search"
        showClear={false}
      />,
    );
    expect(screen.queryByRole('button', { name: /clear search/i })).toBeNull();
  });

  it('renders infoText when provided', () => {
    render(
      <DebouncedSearchInput
        value="q"
        onChange={jest.fn()}
        placeholder="Search"
        infoText={<span>1 result</span>}
      />,
    );
    expect(screen.getByText('1 result')).toBeInTheDocument();
  });
});
