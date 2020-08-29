import React, { Fragment } from "react";

class Home extends React.Component {
  constructor(props) {
    super(props)

    this.state = {
      homes: "homes"
    }
  }

  render() {
    return (
      <Fragment>
        <hr />
        WELCOME PLEASE GO TO THE ACCOUNTS PAGE
      </Fragment>
    );
  }
};

export default Home;
