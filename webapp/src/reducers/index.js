import {combineReducers} from 'redux';

import ActionTypes from '../action_types';

function lastFetchResultByEmail(state = {}, action) {
    switch (action.type) {
    case ActionTypes.RECEIVED_PINGBOARD_INFO: {
        const nextState = {...state};
        nextState[action.email] = action.fetchResult;
        return nextState;
    }
    default:
        return state;
    }
}

export default combineReducers({
    lastFetchResultByEmail,
});
