import { render } from '@testing-library/react';
import { useLockBodyScroll } from './useLockBodyScroll';

function Harness({ locked }: { locked: boolean }) {
  useLockBodyScroll(locked);
  return null;
}

describe('useLockBodyScroll', () => {
  afterEach(() => {
    document.body.style.overflow = '';
    document.body.style.paddingRight = '';
  });

  it('sets overflow hidden while locked and restores on unlock', () => {
    const { rerender, unmount } = render(<Harness locked />);
    expect(document.body.style.overflow).toBe('hidden');

    rerender(<Harness locked={false} />);
    expect(document.body.style.overflow).toBe('');

    unmount();
  });

  it('does nothing when locked is false', () => {
    const prevOverflow = document.body.style.overflow;
    const { unmount } = render(<Harness locked={false} />);
    expect(document.body.style.overflow).toBe(prevOverflow);
    unmount();
  });

  it('keeps body locked until the last nested lock unmounts', () => {
    const { rerender, unmount } = render(
      <>
        <Harness locked />
        <Harness locked />
      </>
    );
    expect(document.body.style.overflow).toBe('hidden');

    rerender(
      <>
        <Harness locked={false} />
        <Harness locked />
      </>
    );
    expect(document.body.style.overflow).toBe('hidden');

    rerender(
      <>
        <Harness locked={false} />
        <Harness locked={false} />
      </>
    );
    expect(document.body.style.overflow).toBe('');

    unmount();
  });
});
