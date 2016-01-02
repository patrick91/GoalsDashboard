# My Goals Application

This is a simple web application that shows my daily/weekly progresses with
my fitness goals.

It is composed by a GO backend, hosted on Google App Engine and a FrontEnd
that will be made with React.

Right now we have one single endpoint `/api/goals` that right now returns
only my daily steps progresses. It is using the FitBit API to get my daily
steps and goal.

# Deploy

    appcfg.py -A patgoals-1169 update . --noauth_local_webserver

# Configuration

Insert the API keys into `/admin/settings` then authenticate with FitBit here: `/fitbit/auth`.
