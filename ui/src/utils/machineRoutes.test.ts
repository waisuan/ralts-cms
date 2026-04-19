import { machineDetailHref } from './machineRoutes';

describe('machineDetailHref', () => {
  it('encodes serial for the machine detail path', () => {
    expect(machineDetailHref('SN-001')).toBe('/machines/SN-001');
  });

  it('encodes unsafe path characters', () => {
    expect(machineDetailHref('SN#1/2')).toBe('/machines/SN%231%2F2');
  });
});
