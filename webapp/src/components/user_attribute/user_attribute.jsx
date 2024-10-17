import React from 'react';
import PropTypes from 'prop-types';

import {describeTenure} from '@/dateutil';
import manifest from '@/manifest';

const {messageHtmlToComponent, formatText} = window.PostUtils;

export default class UserAttribute extends React.PureComponent {
    static propTypes = {
        username: PropTypes.string,
        pingboardInfo: PropTypes.object,
        fetchAndStorePingboardInfo: PropTypes.func.isRequired,
    }

    constructor(props) {
        super(props);
        props.fetchAndStorePingboardInfo(props.username);
    }

    render() {
        const {pingboardInfo} = this.props;
        if (pingboardInfo == null) {
            return null;
        }
        let tenure = '(unknown)';
        if (pingboardInfo.start_year > 0) {
            const localDate = new Date();
            const startDate = new Date(pingboardInfo.start_year, pingboardInfo.start_month - 1, pingboardInfo.start_day);
            tenure = describeTenure(startDate, localDate);
        }
        const description = pingboardInfo.job_title + (pingboardInfo.department ? ` (${pingboardInfo.department})` : '');
        const manager = pingboardInfo.manager ? `@${pingboardInfo.manager}` : '(unknown manager)';

        return (
            <div>
                <div key={`${manifest.id}_job_title`} style={{textWrap: "pretty"}}>
                    {messageHtmlToComponent(`👤 ${description}`)}
                </div>
                <div key={`${manifest.id}_manager`}>
                    {messageHtmlToComponent(formatText(`⬆️ ${manager}`, {atMentions: true, emoticons: false}))}
                </div>
                <div key={`${manifest.id}_start_date`}>
                    {messageHtmlToComponent(`🗓 ${tenure}`)}
                </div>
                <div key={`${manifest.id}_phone`}>
                    {messageHtmlToComponent(`📞 ${pingboardInfo.phone}`)}
                </div>
                <div key={`${manifest.id}_link`}>
                    {messageHtmlToComponent(`↪ <a href=${pingboardInfo.url} target="_blank">Pingboard profile</a>`)}
                </div>
            </div>
        );
    }
}
