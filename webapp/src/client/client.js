import {Client4} from 'mattermost-redux/client';
import {ClientError} from 'mattermost-redux/client/client4';

import {id as pluginId} from '../manifest';

export default class Client {
    getPingboardInfo = async (username = '') => {
        const url = `/plugins/${pluginId}/user?username=` + username;
        const response = await fetch(url, Client4.getOptions({
            method: 'get',
        }));
        if (response.ok) {
            return response.json();
        }
        const text = await response.text();
        throw new ClientError(Client4.url, {
            message: text || '',
            status_code: response.status,
            url,
        });
    };
}
