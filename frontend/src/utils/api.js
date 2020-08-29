import axios from "axios";

export async function getLinkToken(setState) {
    try {
      const response = await axios.post('http://localhost:8000/api/create_link_token', {});
      console.log('Returned data:', response.data.link_token);

      return response.data.link_token
    } catch (e) {
      console.log(`Axios request failed for getting a link token: ${e}`);
    }
}

export async function getTransactions(params) {
    try {
        const response = await axios.post('http://localhost:8000/api/transactions', params);
        return response;
    } catch (e) {
        console.log(`Axios request failed for getting transactions for all accounts: ${e}`);
    }
}

export async function generateAccessToken(params) {
    try {
        const response = await axios.post('http://localhost:8000/api/set_access_token', params);
        return response;
    } catch (e) {
        console.log(`Axios request failed for generating an access token: ${e}`);
    }
}