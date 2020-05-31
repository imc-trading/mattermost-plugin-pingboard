const MONTHS_IN_YEAR = 12;

export function describeTenure(startDate, localDate) {
    if (startDate > localDate) {
        return '';
    }
    let years = localDate.getFullYear() - startDate.getFullYear();
    let months = localDate.getMonth() - startDate.getMonth();
    if (localDate.getDate() < startDate.getDate()) {
        months -= 1;
    }
    if (months < 0) {
        years -= 1;
        months += MONTHS_IN_YEAR;
    }

    if (years === 0 && months === 0) {
        return 'New starter';
    }

    let tenure = '';
    if (years > 0) {
        tenure = years + ' year' + (years === 1 ? '' : 's');
    }
    if (months > 0) {
        if (tenure) {
            tenure += ', ';
        }
        tenure += months + ' month' + (months === 1 ? '' : 's');
    }
    return tenure;
}
