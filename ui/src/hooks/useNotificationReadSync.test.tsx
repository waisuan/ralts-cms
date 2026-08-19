import { renderHook, act } from '@testing-library/react';
import { useNotificationReadSync, NOTIFICATIONS_READ_EVENT } from './useNotificationReadSync';

describe('useNotificationReadSync', () => {
  it('tells the other views, but not the one that announced', () => {
    const bell = jest.fn();
    const inbox = jest.fn();
    const inboxView = renderHook(() => useNotificationReadSync(inbox));
    renderHook(() => useNotificationReadSync(bell));

    act(() => inboxView.result.current());

    expect(bell).toHaveBeenCalledTimes(1);
    // The announcer has already applied the change to its own state.
    expect(inbox).not.toHaveBeenCalled();
  });

  it('runs the handler as it is at the time of the announcement', () => {
    const first = jest.fn();
    const second = jest.fn();
    const listener = renderHook(({ handler }) => useNotificationReadSync(handler), {
      initialProps: { handler: first },
    });
    const announcer = renderHook(() => useNotificationReadSync(jest.fn()));

    listener.rerender({ handler: second });
    act(() => announcer.result.current());

    expect(first).not.toHaveBeenCalled();
    expect(second).toHaveBeenCalledTimes(1);
  });

  it('stops listening once the view is gone', () => {
    const gone = jest.fn();
    const listener = renderHook(() => useNotificationReadSync(gone));
    listener.unmount();

    act(() => {
      window.dispatchEvent(new Event(NOTIFICATIONS_READ_EVENT));
    });

    expect(gone).not.toHaveBeenCalled();
  });
});
