
export function createAccountData(accounts, transactions) {
    var accountsInfo = {};

    for (var i in accounts) {
        var acct = accounts[i]
        accountsInfo[acct.account_id] = {
            name: acct.name,
            balances: acct.balances,
            transactions: [],
            type: acct.subtype
        };
    };

    for (var j in transactions) {
        var tran = transactions[j]
        accountsInfo[tran.account_id].transactions.push({
            amount: tran.amount,
            date: tran.date,
            category: tran.category,
            market: tran.name
        });
    }

    return accountsInfo
}

export function getTotalBankBalance(accounts) {
    var balance = 0.0
    for (var i in accounts) {
        var acct = accounts[i]
        balance += acct.balances.current
    }

    return balance
}