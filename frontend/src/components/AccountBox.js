import React from 'react';

import Button from '@material-ui/core/Button';
import DialogTitle from '@material-ui/core/DialogTitle';
import Dialog from '@material-ui/core/Dialog';

import { getTotalBankBalance } from "../utils/data"

export class AccountBox extends React.Component {
    constructor(props) {
        super(props)

        this.handleOpen = this.handleOpen.bind(this)
        this.handleClose = this.handleClose.bind(this)

        this.state = {
            visible: false
        }
    }

    handleOpen(e) {
        this.setState({
            visible: true
        })
    }
    
    handleClose(e) {
        this.setState({
            visible: false
        })
    }

    render() {
        return (
            <div>
                <Button variant="outlined" color="primary" fullWidth={true} onClick={this.handleOpen}
                    style={{minWidth: "250px", minHeight: "250px"}}
                >
                    <div>
                        <div style={{fontSize: "16px"}}>
                            {this.props.instName}
                        </div>
                        <br />
                        <div>
                            {"Balance: " + getTotalBankBalance(this.props.instData).toFixed(2)}
                            <br />
                            {"Number of Accounts: " + Object.keys(this.props.instData).length}
                        </div>
                    </div>
                </Button>
                <Dialog onClose={this.handleClose} aria-labelledby="simple-dialog-title" open={this.state.visible}>
                <DialogTitle id="simple-dialog-title">Set backup account</DialogTitle>
                    HELLO DIALOG
                </Dialog>
            </div>
        )
    }
}

export default AccountBox;