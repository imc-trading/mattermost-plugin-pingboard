/* eslint-disable no-magic-numbers */
import {describeTenure} from 'dateutil';

test('Tenure descriptions are as expected', () => {
    const startDate = new Date(2010, 6, 10);
    expect(describeTenure(startDate, new Date(2010, 1, 30))).toStrictEqual('');
    expect(describeTenure(startDate, new Date(2010, 6, 10))).toStrictEqual('New starter');
    expect(describeTenure(startDate, new Date(2010, 6, 30))).toStrictEqual('New starter');
    expect(describeTenure(startDate, new Date(2010, 7, 9))).toStrictEqual('New starter');
    expect(describeTenure(startDate, new Date(2010, 7, 10))).toStrictEqual('1 month');
    expect(describeTenure(startDate, new Date(2010, 7, 25))).toStrictEqual('1 month');
    expect(describeTenure(startDate, new Date(2010, 8, 9))).toStrictEqual('1 month');
    expect(describeTenure(startDate, new Date(2010, 8, 10))).toStrictEqual('2 months');
    expect(describeTenure(startDate, new Date(2011, 1, 5))).toStrictEqual('6 months');
    expect(describeTenure(startDate, new Date(2011, 1, 10))).toStrictEqual('7 months');
    expect(describeTenure(startDate, new Date(2011, 6, 9))).toStrictEqual('11 months');
    expect(describeTenure(startDate, new Date(2011, 6, 10))).toStrictEqual('1 year');
    expect(describeTenure(startDate, new Date(2011, 8, 30))).toStrictEqual('1 year, 2 months');
    expect(describeTenure(startDate, new Date(2011, 10, 5))).toStrictEqual('1 year, 3 months');
    expect(describeTenure(startDate, new Date(2011, 10, 11))).toStrictEqual('1 year, 4 months');
    expect(describeTenure(startDate, new Date(2013, 6, 10))).toStrictEqual('3 years');
    expect(describeTenure(startDate, new Date(2099, 2, 15))).toStrictEqual('88 years, 8 months');
});
