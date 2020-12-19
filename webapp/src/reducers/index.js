import {combineReducers} from 'redux';

import ActionTypes from '../action_types';

function lastFetchResultByUsername(state = {}, action) {
    switch (action.type) {
    case ActionTypes.RECEIVED_PINGBOARD_INFO: {
        const nextState = {...state};
        nextState[action.username] = action.fetchResult;
        return nextState;
    }
    default:
        return state;
    }
}

export default combineReducers({
    lastFetchResultByUsername,
});
