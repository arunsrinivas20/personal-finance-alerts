import React from "react";
import { Auth0Context } from "@auth0/auth0-react";
import PlaidConnect from "../components/PlaidConnect";
import { getLinkToken, generateAccessToken, getTransactions, getLinkedAccounts } from "../utils/api"
import { createAccountData } from "../utils/data"
import Box from '@material-ui/core/Box';
import './Accounts.scss'; 

import {AccountBox} from "../components/AccountBox"

class Accounts extends React.Component {
  static contextType = Auth0Context;

  constructor(props) {
    super(props)

    this.handleOnSuccess = this.handleOnSuccess.bind(this);
    this.handleOnExit = this.handleOnExit.bind(this);
    this.renderAccounts = this.renderAccounts.bind(this);

    this.state = {
      linkToken: "",
      bankData: {}
    }
  }

  async componentDidMount() {
    console.log("MOUNT")
    const {user} = this.context;
    
    await getLinkedAccounts(user.name)
    .then(response => {
        console.log(response)
        this.setState({
            bankData: response
        })
    })
    .catch(e => {
        console.log(e)
    })
    
  }

  async handleOnSuccess (public_token, metadata) {
    const {user} = this.context;
    
    console.log(public_token)
    console.log(metadata)
    
    var params = {
        "public_token": public_token, 
        "email": user.name, 
        "institution_id": metadata.institution.institution_id,
        "institution_name": metadata.institution.name
    }

    await generateAccessToken(params)
    .then(async response => {
        console.log(response)

        params = {
            "email": user.name, 
            "institution_id": metadata.institution.institution_id,
            "institution_name": metadata.institution.name
        }

        await getTransactions(params)
        .then(response => {
            console.log(response);

            var bankAccountsData = createAccountData(response.data.accounts, response.data.transactions)
            var currBankData = this.state.bankData

            currBankData[metadata.institution.name] = bankAccountsData

            this.setState({
                bankData: currBankData
            })

            this.props.handleUpdateAcctInfo({
                "public_token": public_token, 
                "metadata": metadata
            });
        })
        .catch(error => {
            console.log(error)
        })
    })
    .catch(error => {
        console.log(error)
    })
  }

  handleOnExit(){}

  renderAccounts() {
    var res = []

    for (var inst in this.state.bankData) {
        res.push(
            <div key={inst} className="equalHMV eq">
                <AccountBox instName={inst} instData={this.state.bankData[inst]}/>
            </div>
        )
    }

    return res
  }

  render() {
    const {isAuthenticated} = this.context

    if (this.state.linkToken === "") {
      getLinkToken()
      .then(response => {
          this.setState({
            linkToken: response
          })
      })
      .catch(error => {
          console.log(error)
      })
    }
    return (
      <div>
        {isAuthenticated ? (
          this.state.linkToken !== "" ? (
              <div style={{paddingBottom: "20px", paddingLeft: "10px"}}>
                  <PlaidConnect 
                    handleOnSuccess={this.handleOnSuccess} 
                    handleOnExit={this.handleOnExit} 
                    linkToken={this.state.linkToken}
                />
              </div>
            ) : "\n  NO LINK"
          ) : "\nYou are not logged in"
        }
        <div className="equalHMVWrap eqWrap">
            {this.renderAccounts()}
        </div>
      </div>
    );
  }
};

export default Accounts;
