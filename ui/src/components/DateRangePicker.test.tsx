import { render, screen } from '@testing-library/react';
import DateRangePicker from './DateRangePicker';

describe('DateRangePicker', () => {
  it('does not nest a native button inside the trigger button', () => {
    const noop = () => {};
    render(
      <DateRangePicker
        label="PPM date"
        value={{ from: '2024-01-01', to: '2024-01-31' }}
        onChange={noop}
      />
    );

    const trigger = screen.getByRole('button', { name: /Jan 1, 2024/ });
    expect(trigger.querySelector('button')).toBeNull();
  });
});
