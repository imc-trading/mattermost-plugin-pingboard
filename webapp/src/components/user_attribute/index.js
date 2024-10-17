import {connect} from 'react-redux';
import {bindActionCreators} from 'redux';

import manifest from '@/manifest';
import {fetchAndStorePingboardInfo} from '@/actions';

import UserAttribute from './user_attribute.jsx';

const REDUCER = `plugins-${manifest.id}`;

function mapStateToProps(state, ownProps) {
    let username;
    let pingboardInfo;
    if (ownProps.user) {
        username = ownProps.user.username;
        const lastFetchResult = state[REDUCER].lastFetchResultByUsername[username];
        if (lastFetchResult) {
            pingboardInfo = lastFetchResult.pingboardInfo;
        }
    }

    return {
        username,
        pingboardInfo,
    };
}

function mapDispatchToProps(dispatch) {
    return bindActionCreators({
        fetchAndStorePingboardInfo,
    }, dispatch);
}

export default connect(mapStateToProps, mapDispatchToProps)(UserAttribute);
