import Client from '../client';
import ActionTypes from '../action_types';

export function fetchAndStorePingboardInfo(email = '') {
    return async (dispatch) => {
        let pingboardInfo;
        try {
            pingboardInfo = await Client.getPingboardInfo(email);
        } catch (error) {
            if (error.status === 404) {
                return dispatch({
                    type: ActionTypes.RECEIVED_PINGBOARD_INFO,
                    email,
                    fetchResult: {
                        pingboardInfo: null,
                    },
                });
            }
            throw error;
        }

        return dispatch({
            type: ActionTypes.RECEIVED_PINGBOARD_INFO,
            email,
            fetchResult: {
                pingboardInfo,
            },
        });
    };
}
