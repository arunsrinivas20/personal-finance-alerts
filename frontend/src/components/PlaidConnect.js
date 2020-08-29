import React from "react";
import { PlaidLink } from "react-plaid-link";

const PlaidConnect = (props) => {
    return (
        <div className="nav-container">
            <PlaidLink
                clientName="random"
                env="sandbox"
                product={["auth"]}
                onExit={props.handleOnExit}
                onSuccess={props.handleOnSuccess}
                apiVersion={'v2'}
                token={props.linkToken}
            >
                Add an Account!
            </PlaidLink>
        </div>
    )
}

export default PlaidConnect;