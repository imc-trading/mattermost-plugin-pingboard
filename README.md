# Mattermost Pingboard Plugin

This plugin collects information on users from Pingboard and displays
it on the user popover card.

## Features

If data is found for the user, the user's popover card is extended with:
* Job title and department
* Years/months since start date
* Phone number
* @-mention for manager (if manager was also found as a mattermost user)
* Link to user's Pingboard profile

![Screenshot](screenshot.png)

## Pre-requisites

The plugin matches users based on email address, so this must match
in Pingboard and Mattermost. The email addresses are first normalised
by stripping out all characters except letters, digits and dots, and
then compared as lowercase - so `a-Strange.Email+address@somewhere.com`
will match `astrange.emailaddress@somewhere.com`.

Create a client ID for Pingboard with read-only access to user data
and note the client ID and client secret.

**Note**: information for any Mattermost user that exists in Pingboard
can then be seen by all users. If access to that information needs
to be managed, the plugin may not be suitable.

## Configuration

The Pingboard API client ID and secret should be supplied in the plugin
config via the System Console.

If preferred, the client secret can be left blank, and instead provided by
setting the environment variable `MM_PLUGIN_PINGBOARD_CLIENT_SECRET` for the
mattermost server.

## Implementation notes

* Pingboard is queried for company information (for inserting sub-domain into pingboard link URLs),
  and all known users. The first valid group listed under the user's departments is also looked up
  to get the department name.
* Pingboard users are then matched by email address against mattermost users. The email address
  match ignores all characters except letters, digits and dots, and compares in lowercase.
* The resulting data is held in memory in the server plugin and fetched again every 6 hours, or
  when a new user is created.
* The client looks up information for a user by username via the plugin's internal http endpoint.
  The information is retrieved from the server every time a popover is created.
