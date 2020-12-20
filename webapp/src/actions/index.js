import Client from '../client';
import ActionTypes from '../action_types';

const HTTP_NOT_FOUND = 404;

export function fetchAndStorePingboardInfo(username = '') {
    return async (dispatch) => {
        let pingboardInfo;
        try {
            pingboardInfo = await Client.getPingboardInfo(username);
        } catch (error) {
            if (error.status === HTTP_NOT_FOUND) {
                return dispatch({
                    type: ActionTypes.RECEIVED_PINGBOARD_INFO,
                    username,
                    fetchResult: {
                        pingboardInfo: null,
                    },
                });
            }
            throw error;
        }

        return dispatch({
            type: ActionTypes.RECEIVED_PINGBOARD_INFO,
            username,
            fetchResult: {
                pingboardInfo,
            },
        });
    };
}
