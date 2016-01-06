import React from 'react';


class StepsGoal extends React.Component {
    render() {
        return <div className="steps steps--goal-reached">
            <div></div>

            <div>
                <div className="steps__goal">13.000</div>
                <div className="steps__current steps__current--goal-reached"></div>
            </div>

            <div className="steps__caption">steps</div>
        </div>;
    }
}

export default StepsGoal;
