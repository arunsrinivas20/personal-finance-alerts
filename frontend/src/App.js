import React from "react";
import { Router, Route, Switch } from "react-router-dom";
import { Container } from "reactstrap";

import Loading from "./components/Loading";
import NavBar from "./components/NavBar";
import Footer from "./components/Footer";
import Home from "./views/Home";
import Profile from "./views/Profile";
import Accounts from "./views/Accounts";
import { Auth0Context } from "@auth0/auth0-react";
import history from "./utils/history";

// styles
import "./App.css";

// fontawesome
import initFontAwesome from "./utils/initFontAwesome";
initFontAwesome();

class App extends React.Component {
  static contextType = Auth0Context

  constructor(props) {
    super(props)

    this.handleUpdateAcctInfo = this.handleUpdateAcctInfo.bind(this)

    this.state = {
      accountInfo: []
    }
  }

  handleUpdateAcctInfo(newAcct) {
    this.setState({
      accountInfo: this.state.accountInfo.concat(newAcct)
    })
  }

  render() {
    const { isLoading, error } = this.context;

    if (error) {
      return <div>Oops... {error.message}</div>;
    }

    if (isLoading) {
      return <Loading />;
    }

    return (
      <Router history={history}>
        <div id="app" className="d-flex flex-column h-100">
          <NavBar />
          <Container className="flex-grow-1 mt-5">
            <Switch>
              <Route 
                path="/" 
                exact 
                render={(props) => (
                  <Home {...props} />
                )}
              />
              <Route path="/profile" component={Profile} />
              <Route 
                path="/accounts" 
                render={(props) => (
                  <Accounts {...props} accountInfo={this.state.accountInfo} handleUpdateAcctInfo={this.handleUpdateAcctInfo} />
                )}
              />
            </Switch>
          </Container>
          <Footer />
        </div>
      </Router>
    );
  }
};

export default App;
