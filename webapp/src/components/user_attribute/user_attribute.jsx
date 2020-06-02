import React from 'react';
import PropTypes from 'prop-types';

import {id as pluginId} from '../../manifest';
const {messageHtmlToComponent} = window.PostUtils;

export default class UserAttribute extends React.PureComponent {
    static propTypes = {
        email: PropTypes.string,
        pingboardInfo: PropTypes.object,
        fetchAndStorePingboardInfo: PropTypes.func.isRequired,
    }

    constructor(props) {
        super(props);
        props.fetchAndStorePingboardInfo(props.email);
    }

    render() {
        const {pingboardInfo} = this.props;
        if (pingboardInfo == null) {
            return null;
        }
        const description = pingboardInfo.job_title + (pingboardInfo.department ? ` (${pingboardInfo.department})` : '');
        return (
            <div>
                <div key={`${pluginId}_job_title`}>
                    {messageHtmlToComponent(
                        `<a href=${pingboardInfo.url} target="_blank">👤 ${description}</a>`)}
                </div>
                <div key={`${pluginId}_start_date`}>
                    {messageHtmlToComponent(`🎂 ${pingboardInfo.start_date}`)}
                </div>
                <div key={`${pluginId}_phone`}>
                    {messageHtmlToComponent(`📞 ${pingboardInfo.phone}`)}
                </div>
            </div>
        );
    }
}
