# Wallet


:moneybag:Wallet service API written in Golang. This api comes fully loaded with the following features
- Authentication and Authorization :closed_lock_with_key:
- Concurrency
- Database Transactions using Postgres


## Project Overview
On a very high level this is an added service to a webapp that allows a user fund his/her account and make transactions from the wallet till it is depleted.
It gives the user the opportunity to send to other verified users money, allows for payment of items and provides efficient tracking of transactions and balance.


#### Authentication
- Register with email, username and password
- Login with username and password
- Use of token authentication :key:
- Return secure cookies on login and refresh of token


#### Transactions
- Fund wallet with paystack :pound:
- Verify transactions from paystack
- Update balance when transaction is carried out. Use of database transactions in case of errors :star2:
- Perform wallet to wallet transfers.


-----------------------------------------------------------------------------------------------------------------------------------------------------
This was built without any framework and is currently hosted with heroku. For further enquiries please send me a mail aloziekelechi17@gmail.com :grin:

Find something you like, leave a like please :grin:, I like the attention :wink:
