import {connect} from 'react-redux';
import {bindActionCreators} from 'redux';

import {id as pluginId} from '../../manifest';
import {fetchAndStorePingboardInfo} from '../../actions';

import UserAttribute from './user_attribute.jsx';

const REDUCER = `plugins-${pluginId}`;

function mapStateToProps(state, ownProps) {
    let email;
    let pingboardInfo;
    if (ownProps.user) {
        email = ownProps.user.email;
        const lastFetchResult = state[REDUCER].lastFetchResultByEmail[email];
        if (lastFetchResult) {
            pingboardInfo = lastFetchResult.pingboardInfo;
        }
    }

    return {
        email,
        pingboardInfo,
    };
}

function mapDispatchToProps(dispatch) {
    return bindActionCreators({
        fetchAndStorePingboardInfo,
    }, dispatch);
}

export default connect(mapStateToProps, mapDispatchToProps)(UserAttribute);
