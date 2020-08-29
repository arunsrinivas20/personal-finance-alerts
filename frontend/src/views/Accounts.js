import React from "react";
import { Auth0Context } from "@auth0/auth0-react";
import PlaidConnect from "../components/PlaidConnect";
import { getLinkToken, generateAccessToken, getTransactions } from "../utils/api"
import { createAccountData } from "../utils/data"

class Accounts extends React.Component {
  static contextType = Auth0Context;

  constructor(props) {
    super(props)

    this.handleOnSuccess = this.handleOnSuccess.bind(this);
    this.handleOnExit = this.handleOnExit.bind(this);

    this.state = {
      linkToken: "",
      bankData: {}
    }
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
      <div className="text-center hero my-5">
        {isAuthenticated ? (
          this.state.linkToken !== "" ? (
              <PlaidConnect 
                handleOnSuccess={this.handleOnSuccess} 
                handleOnExit={this.handleOnExit} 
                linkToken={this.state.linkToken}
              />
            ) : "\n  NO LINK"
          ) : "\nYou are not logged in"
        }
      </div>
    );
  }
};

export default Accounts;
