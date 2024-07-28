package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	mockdb "github.com/taylordurden/go-simple-bank/db/mock"
	db "github.com/taylordurden/go-simple-bank/db/sqlc"
	"github.com/taylordurden/go-simple-bank/token"
	"github.com/taylordurden/go-simple-bank/util"
)

func TestCreateTransferAPI(t *testing.T) {
	var amount int64
	fromUser, _ := randomUser(t)
	toUser, _ := randomUser(t)
	toUser1, _ := randomUser(t)
	fromAccount := randomAccount(fromUser.Username)
	fromAccount.Currency = util.USD
	toAccount := randomAccount(toUser.Username)
	toAccount.Currency = util.USD
	toAccount1 := randomAccount(toUser1.Username)
	toAccount1.Currency = util.RMB
	amount = 10

	transfer := db.Transfer{
		ID:            1,
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
		Amount:        amount,
		CreatedAt:     time.Now(),
	}

	txResult := db.TransferTxResult{
		Transfer:    transfer,
		FromAccount: fromAccount,
		ToAccount:   toAccount,
		FromEntry: db.Entry{
			ID:        1,
			Amount:    -amount,
			AccountID: fromAccount.ID,
			CreatedAt: time.Now(),
		},
		ToEntry: db.Entry{
			ID:        2,
			Amount:    amount,
			AccountID: toAccount.ID,
			CreatedAt: time.Now(),
		},
	}

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenAuth token.Authenticator)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenAuth token.Authenticator) {
				addAuthorization(t, request, tokenAuth, authorizationTypeBearer, fromUser.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.TransferTxParams{
					FromAccountID: fromAccount.ID,
					ToAccountID:   toAccount.ID,
					Amount:        int64(amount),
				}

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).
					Times(1).
					Return(toAccount, nil)
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(txResult, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireResponseBodyEqualTransferResult(t, *recorder.Body, txResult)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenAuth token.Authenticator) {
				addAuthorization(t, request, tokenAuth, authorizationTypeBearer, fromUser.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).Times(1).Return(toAccount, nil)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(1).Return(db.TransferTxResult{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "BadRequestInvalidCurrencyToAccount",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount1.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenAuth token.Authenticator) {
				addAuthorization(t, request, tokenAuth, authorizationTypeBearer, fromUser.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount1.ID)).Times(1).Return(toAccount1, nil)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "NotFoundToAccount",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount1.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenAuth token.Authenticator) {
				addAuthorization(t, request, tokenAuth, authorizationTypeBearer, fromUser.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount1.ID)).Times(1).Return(db.Account{}, sql.ErrNoRows)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "FromAccountCurrencyMismatch",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        util.RMB,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenAuth token.Authenticator) {
				addAuthorization(t, request, tokenAuth, authorizationTypeBearer, fromUser.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(fromAccount, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "ToAccountCurrencyMismatch",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount1.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenAuth token.Authenticator) {
				addAuthorization(t, request, tokenAuth, authorizationTypeBearer, fromUser.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(toAccount1.ID)).Times(1).Return(toAccount1, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "GetAccountInternalError",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount1.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenAuth token.Authenticator) {
				addAuthorization(t, request, tokenAuth, authorizationTypeBearer, fromUser.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).Times(1).Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "BadRequest",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount1.ID,
				"amount":          amount,
				"currency":        "JAP",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenAuth token.Authenticator) {
				addAuthorization(t, request, tokenAuth, authorizationTypeBearer, fromUser.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			// build stubs
			tc.buildStubs(store)

			// start test server and send request
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/transfers"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenAuth)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func requireResponseBodyEqualTransferResult(t *testing.T, resBody bytes.Buffer, txResult db.TransferTxResult) {
	data, err := io.ReadAll(&resBody)
	require.NoError(t, err)

	var resTransferResult db.TransferTxResult
	err = json.Unmarshal(data, &resTransferResult)
	require.NoError(t, err)

	// require.Equal(t, txResult, resTransferResult)
	util.AreStructsEqual(txResult, resTransferResult)
}
