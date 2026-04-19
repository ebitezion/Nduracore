package transaction

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTxResult_String(t *testing.T) {
	tests := []struct {
		name     string
		txResult TxResult
		expected string
	}{
		// Tec Codes
		{
			name:     "TecAMM_ACCOUNT",
			txResult: TecAMM_ACCOUNT,
			expected: "tecAMM_ACCOUNT",
		},
		{
			name:     "TecAMM_UNFUNDED",
			txResult: TecAMM_UNFUNDED,
			expected: "tecAMM_UNFUNDED",
		},
		{
			name:     "TecAMM_BALANCE",
			txResult: TecAMM_BALANCE,
			expected: "tecAMM_BALANCE",
		},
		{
			name:     "TecAMM_EMPTY",
			txResult: TecAMM_EMPTY,
			expected: "tecAMM_EMPTY",
		},
		{
			name:     "TecAMM_FAILED",
			txResult: TecAMM_FAILED,
			expected: "tecAMM_FAILED",
		},
		{
			name:     "TecAMM_INVALID_TOKENS",
			txResult: TecAMM_INVALID_TOKENS,
			expected: "tecAMM_INVALID_TOKENS",
		},
		{
			name:     "TecAMM_NOT_EMPTY",
			txResult: TecAMM_NOT_EMPTY,
			expected: "tecAMM_NOT_EMPTY",
		},
		{
			name:     "TecCANT_ACCEPT_OWN_NFTOKEN_OFFER",
			txResult: TecCANT_ACCEPT_OWN_NFTOKEN_OFFER,
			expected: "tecCANT_ACCEPT_OWN_NFTOKEN_OFFER",
		},
		{
			name:     "TecCLAIM",
			txResult: TecCLAIM,
			expected: "tecCLAIM",
		},
		{
			name:     "TecCRYPTOCONDITION_ERROR",
			txResult: TecCRYPTOCONDITION_ERROR,
			expected: "tecCRYPTOCONDITION_ERROR",
		},
		{
			name:     "TecDIR_FULL",
			txResult: TecDIR_FULL,
			expected: "tecDIR_FULL",
		},
		{
			name:     "TecDUPLICATE",
			txResult: TecDUPLICATE,
			expected: "tecDUPLICATE",
		},
		{
			name:     "TecDST_TAG_NEEDED",
			txResult: TecDST_TAG_NEEDED,
			expected: "tecDST_TAG_NEEDED",
		},
		{
			name:     "TecEMPTY_DID",
			txResult: TecEMPTY_DID,
			expected: "tecEMPTY_DID",
		},
		{
			name:     "TecEXPIRED",
			txResult: TecEXPIRED,
			expected: "tecEXPIRED",
		},
		{
			name:     "TecFAILED_PROCESSING",
			txResult: TecFAILED_PROCESSING,
			expected: "tecFAILED_PROCESSING",
		},
		{
			name:     "TecFROZEN",
			txResult: TecFROZEN,
			expected: "tecFROZEN",
		},
		{
			name:     "TecHAS_OBLIGATIONS",
			txResult: TecHAS_OBLIGATIONS,
			expected: "tecHAS_OBLIGATIONS",
		},
		{
			name:     "TecINSUF_RESERVE_LINE",
			txResult: TecINSUF_RESERVE_LINE,
			expected: "tecINSUF_RESERVE_LINE",
		},
		{
			name:     "TecINSUF_RESERVE_OFFER",
			txResult: TecINSUF_RESERVE_OFFER,
			expected: "tecINSUF_RESERVE_OFFER",
		},
		{
			name:     "TecINSUFF_FEE",
			txResult: TecINSUFF_FEE,
			expected: "tecINSUFF_FEE",
		},
		{
			name:     "TecINSUFFICIENT_FUNDS",
			txResult: TecINSUFFICIENT_FUNDS,
			expected: "tecINSUFFICIENT_FUNDS",
		},
		{
			name:     "TecINSUFFICIENT_PAYMENT",
			txResult: TecINSUFFICIENT_PAYMENT,
			expected: "tecINSUFFICIENT_PAYMENT",
		},
		{
			name:     "TecINSUFFICIENT_RESERVE",
			txResult: TecINSUFFICIENT_RESERVE,
			expected: "tecINSUFFICIENT_RESERVE",
		},
		{
			name:     "TecINTERNAL",
			txResult: TecINTERNAL,
			expected: "tecINTERNAL",
		},
		{
			name:     "TecINVARIANT_FAILED",
			txResult: TecINVARIANT_FAILED,
			expected: "tecINVARIANT_FAILED",
		},
		{
			name:     "TecKILLED",
			txResult: TecKILLED,
			expected: "tecKILLED",
		},
		{
			name:     "TecMAX_SEQUENCE_REACHED",
			txResult: TecMAX_SEQUENCE_REACHED,
			expected: "tecMAX_SEQUENCE_REACHED",
		},
		{
			name:     "TecNEED_MASTER_KEY",
			txResult: TecNEED_MASTER_KEY,
			expected: "tecNEED_MASTER_KEY",
		},
		{
			name:     "TecNFTOKEN_BUY_SELL_MISMATCH",
			txResult: TecNFTOKEN_BUY_SELL_MISMATCH,
			expected: "tecNFTOKEN_BUY_SELL_MISMATCH",
		},
		{
			name:     "TecNFTOKEN_OFFER_TYPE_MISMATCH",
			txResult: TecNFTOKEN_OFFER_TYPE_MISMATCH,
			expected: "tecNFTOKEN_OFFER_TYPE_MISMATCH",
		},
		{
			name:     "TecNO_ALTERNATIVE_KEY",
			txResult: TecNO_ALTERNATIVE_KEY,
			expected: "tecNO_ALTERNATIVE_KEY",
		},
		{
			name:     "TecNO_AUTH",
			txResult: TecNO_AUTH,
			expected: "tecNO_AUTH",
		},
		{
			name:     "TecNO_DST",
			txResult: TecNO_DST,
			expected: "tecNO_DST",
		},
		{
			name:     "TecNO_DST_INSUF_XRP",
			txResult: TecNO_DST_INSUF_XRP,
			expected: "tecNO_DST_INSUF_XRP",
		},
		{
			name:     "TecNO_ENTRY",
			txResult: TecNO_ENTRY,
			expected: "tecNO_ENTRY",
		},
		{
			name:     "TecNO_ISSUER",
			txResult: TecNO_ISSUER,
			expected: "tecNO_ISSUER",
		},
		{
			name:     "TecNO_LINE",
			txResult: TecNO_LINE,
			expected: "tecNO_LINE",
		},
		{
			name:     "TecNO_LINE_INSUF_RESERVE",
			txResult: TecNO_LINE_INSUF_RESERVE,
			expected: "tecNO_LINE_INSUF_RESERVE",
		},
		{
			name:     "TecNO_LINE_REDUNDANT",
			txResult: TecNO_LINE_REDUNDANT,
			expected: "tecNO_LINE_REDUNDANT",
		},
		{
			name:     "TecNO_PERMISSION",
			txResult: TecNO_PERMISSION,
			expected: "tecNO_PERMISSION",
		},
		{
			name:     "TecNO_REGULAR_KEY",
			txResult: TecNO_REGULAR_KEY,
			expected: "tecNO_REGULAR_KEY",
		},
		{
			name:     "TecNO_SUITABLE_NFTOKEN_PAGE",
			txResult: TecNO_SUITABLE_NFTOKEN_PAGE,
			expected: "tecNO_SUITABLE_NFTOKEN_PAGE",
		},
		{
			name:     "TecNO_TARGET",
			txResult: TecNO_TARGET,
			expected: "tecNO_TARGET",
		},
		{
			name:     "TecOBJECT_NOT_FOUND",
			txResult: TecOBJECT_NOT_FOUND,
			expected: "tecOBJECT_NOT_FOUND",
		},
		{
			name:     "TecOVERSIZE",
			txResult: TecOVERSIZE,
			expected: "tecOVERSIZE",
		},
		{
			name:     "TecOWNERS",
			txResult: TecOWNERS,
			expected: "tecOWNERS",
		},
		{
			name:     "TecPATH_DRY",
			txResult: TecPATH_DRY,
			expected: "tecPATH_DRY",
		},
		{
			name:     "TecPATH_PARTIAL",
			txResult: TecPATH_PARTIAL,
			expected: "tecPATH_PARTIAL",
		},
		{
			name:     "TecTOO_SOON",
			txResult: TecTOO_SOON,
			expected: "tecTOO_SOON",
		},
		{
			name:     "TecUNFUNDED",
			txResult: TecUNFUNDED,
			expected: "tecUNFUNDED",
		},
		{
			name:     "TecUNFUNDED_ADD",
			txResult: TecUNFUNDED_ADD,
			expected: "tecUNFUNDED_ADD",
		},
		{
			name:     "TecUNFUNDED_PAYMENT",
			txResult: TecUNFUNDED_PAYMENT,
			expected: "tecUNFUNDED_PAYMENT",
		},
		{
			name:     "TecUNFUNDED_OFFER",
			txResult: TecUNFUNDED_OFFER,
			expected: "tecUNFUNDED_OFFER",
		},

		// Tef Codes
		{
			name:     "TefALREADY",
			txResult: TefALREADY,
			expected: "tefALREADY",
		},
		{
			name:     "TefBAD_ADD_AUTH",
			txResult: TefBAD_ADD_AUTH,
			expected: "tefBAD_ADD_AUTH",
		},
		{
			name:     "TefBAD_AUTH",
			txResult: TefBAD_AUTH,
			expected: "tefBAD_AUTH",
		},
		{
			name:     "TefBAD_AUTH_MASTER",
			txResult: TefBAD_AUTH_MASTER,
			expected: "tefBAD_AUTH_MASTER",
		},
		{
			name:     "TefBAD_LEDGER",
			txResult: TefBAD_LEDGER,
			expected: "tefBAD_LEDGER",
		},
		{
			name:     "TefBAD_QUORUM",
			txResult: TefBAD_QUORUM,
			expected: "tefBAD_QUORUM",
		},
		{
			name:     "TefBAD_SIGNATURE",
			txResult: TefBAD_SIGNATURE,
			expected: "tefBAD_SIGNATURE",
		},
		{
			name:     "TefCREATED",
			txResult: TefCREATED,
			expected: "tefCREATED",
		},
		{
			name:     "TefEXCEPTION",
			txResult: TefEXCEPTION,
			expected: "tefEXCEPTION",
		},
		{
			name:     "TefFAILURE",
			txResult: TefFAILURE,
			expected: "tefFAILURE",
		},
		{
			name:     "TefINTERNAL",
			txResult: TefINTERNAL,
			expected: "tefINTERNAL",
		},
		{
			name:     "TefINVARIANT_FAILED",
			txResult: TefINVARIANT_FAILED,
			expected: "tefINVARIANT_FAILED",
		},
		{
			name:     "TefMASTER_DISABLED",
			txResult: TefMASTER_DISABLED,
			expected: "tefMASTER_DISABLED",
		},
		{
			name:     "TefMAX_LEDGER",
			txResult: TefMAX_LEDGER,
			expected: "tefMAX_LEDGER",
		},
		{
			name:     "TefNFTOKEN_IS_NOT_TRANSFERABLE",
			txResult: TefNFTOKEN_IS_NOT_TRANSFERABLE,
			expected: "tefNFTOKEN_IS_NOT_TRANSFERABLE",
		},
		{
			name:     "TefNO_AUTH_REQUIRED",
			txResult: TefNO_AUTH_REQUIRED,
			expected: "tefNO_AUTH_REQUIRED",
		},
		{
			name:     "TefNO_TICKET",
			txResult: TefNO_TICKET,
			expected: "tefNO_TICKET",
		},
		{
			name:     "TefNOT_MULTI_SIGNING",
			txResult: TefNOT_MULTI_SIGNING,
			expected: "tefNOT_MULTI_SIGNING",
		},
		{
			name:     "TefPAST_SEQ",
			txResult: TefPAST_SEQ,
			expected: "tefPAST_SEQ",
		},
		{
			name:     "TefTOO_BIG",
			txResult: TefTOO_BIG,
			expected: "tefTOO_BIG",
		},
		{
			name:     "TefWRONG_PRIOR",
			txResult: TefWRONG_PRIOR,
			expected: "tefWRONG_PRIOR",
		},
		// Tel Codes
		{
			name:     "TelBAD_DOMAIN",
			txResult: TelBAD_DOMAIN,
			expected: "telBAD_DOMAIN",
		},
		{
			name:     "TelBAD_PATH_COUNT",
			txResult: TelBAD_PATH_COUNT,
			expected: "telBAD_PATH_COUNT",
		},
		{
			name:     "TelBAD_PUBLIC_KEY",
			txResult: TelBAD_PUBLIC_KEY,
			expected: "telBAD_PUBLIC_KEY",
		},
		{
			name:     "TelCAN_NOT_QUEUE",
			txResult: TelCAN_NOT_QUEUE,
			expected: "telCAN_NOT_QUEUE",
		},
		{
			name:     "TelCAN_NOT_QUEUE_BALANCE",
			txResult: TelCAN_NOT_QUEUE_BALANCE,
			expected: "telCAN_NOT_QUEUE_BALANCE",
		},
		{
			name:     "TelCAN_NOT_QUEUE_BLOCKS",
			txResult: TelCAN_NOT_QUEUE_BLOCKS,
			expected: "telCAN_NOT_QUEUE_BLOCKS",
		},
		{
			name:     "TelCAN_NOT_QUEUE_BLOCKED",
			txResult: TelCAN_NOT_QUEUE_BLOCKED,
			expected: "telCAN_NOT_QUEUE_BLOCKED",
		},
		{
			name:     "TelCAN_NOT_QUEUE_FEE",
			txResult: TelCAN_NOT_QUEUE_FEE,
			expected: "telCAN_NOT_QUEUE_FEE",
		},
		{
			name:     "TelCAN_NOT_QUEUE_FULL",
			txResult: TelCAN_NOT_QUEUE_FULL,
			expected: "telCAN_NOT_QUEUE_FULL",
		},
		{
			name:     "TelFAILED_PROCESSING",
			txResult: TelFAILED_PROCESSING,
			expected: "telFAILED_PROCESSING",
		},
		{
			name:     "TelINSUF_FEE_P",
			txResult: TelINSUF_FEE_P,
			expected: "telINSUF_FEE_P",
		},
		{
			name:     "TelLOCAL_ERROR",
			txResult: TelLOCAL_ERROR,
			expected: "telLOCAL_ERROR",
		},
		{
			name:     "TelNETWORK_ID_MAKES_TX_NON_CANONICAL",
			txResult: TelNETWORK_ID_MAKES_TX_NON_CANONICAL,
			expected: "telNETWORK_ID_MAKES_TX_NON_CANONICAL",
		},
		{
			name:     "TelNO_DST_PARTIAL",
			txResult: TelNO_DST_PARTIAL,
			expected: "telNO_DST_PARTIAL",
		},
		{
			name:     "TelREQUIRES_NETWORK_ID",
			txResult: TelREQUIRES_NETWORK_ID,
			expected: "telREQUIRES_NETWORK_ID",
		},
		{
			name:     "TelWRONG_NETWORK",
			txResult: TelWRONG_NETWORK,
			expected: "telWRONG_NETWORK",
		},

		// Tem Codes
		{
			name:     "TemBAD_AMM_TOKENS",
			txResult: TemBAD_AMM_TOKENS,
			expected: "temBAD_AMM_TOKENS",
		},
		{
			name:     "TemBAD_AMOUNT",
			txResult: TemBAD_AMOUNT,
			expected: "temBAD_AMOUNT",
		},
		{
			name:     "TemBAD_AUTH_MASTER",
			txResult: TemBAD_AUTH_MASTER,
			expected: "temBAD_AUTH_MASTER",
		},
		{
			name:     "TemBAD_CURRENCY",
			txResult: TemBAD_CURRENCY,
			expected: "temBAD_CURRENCY",
		},
		{
			name:     "TemBAD_EXPIRATION",
			txResult: TemBAD_EXPIRATION,
			expected: "temBAD_EXPIRATION",
		},
		{
			name:     "TemBAD_FEE",
			txResult: TemBAD_FEE,
			expected: "temBAD_FEE",
		},
		{
			name:     "TemBAD_ISSUER",
			txResult: TemBAD_ISSUER,
			expected: "temBAD_ISSUER",
		},
		{
			name:     "TemBAD_LIMIT",
			txResult: TemBAD_LIMIT,
			expected: "temBAD_LIMIT",
		},
		{
			name:     "TemBAD_NFTOKEN_TRANSFER_FEE",
			txResult: TemBAD_NFTOKEN_TRANSFER_FEE,
			expected: "temBAD_NFTOKEN_TRANSFER_FEE",
		},
		{
			name:     "TemBAD_OFFER",
			txResult: TemBAD_OFFER,
			expected: "temBAD_OFFER",
		},
		{
			name:     "TemBAD_PATH",
			txResult: TemBAD_PATH,
			expected: "temBAD_PATH",
		},
		{
			name:     "TemBAD_PATH_LOOP",
			txResult: TemBAD_PATH_LOOP,
			expected: "temBAD_PATH_LOOP",
		},
		{
			name:     "TemBAD_SEND_XRP_LIMIT",
			txResult: TemBAD_SEND_XRP_LIMIT,
			expected: "temBAD_SEND_XRP_LIMIT",
		},
		{
			name:     "TemBAD_SEND_XRP_MAX",
			txResult: TemBAD_SEND_XRP_MAX,
			expected: "temBAD_SEND_XRP_MAX",
		},
		{
			name:     "TemBAD_SEND_XRP_NO_DIRECT",
			txResult: TemBAD_SEND_XRP_NO_DIRECT,
			expected: "temBAD_SEND_XRP_NO_DIRECT",
		},
		{
			name:     "TemBAD_SEND_XRP_PARTIAL",
			txResult: TemBAD_SEND_XRP_PARTIAL,
			expected: "temBAD_SEND_XRP_PARTIAL",
		},
		{
			name:     "TemBAD_SEND_XRP_PATHS",
			txResult: TemBAD_SEND_XRP_PATHS,
			expected: "temBAD_SEND_XRP_PATHS",
		},
		{
			name:     "TemBAD_SEQUENCE",
			txResult: TemBAD_SEQUENCE,
			expected: "temBAD_SEQUENCE",
		},
		{
			name:     "TemBAD_SIGNATURE",
			txResult: TemBAD_SIGNATURE,
			expected: "temBAD_SIGNATURE",
		},
		{
			name:     "TemBAD_SRC_ACCOUNT",
			txResult: TemBAD_SRC_ACCOUNT,
			expected: "temBAD_SRC_ACCOUNT",
		},
		{
			name:     "TemBAD_TRANSFER_RATE",
			txResult: TemBAD_TRANSFER_RATE,
			expected: "temBAD_TRANSFER_RATE",
		},
		{
			name:     "TemCANNOT_PREAUTH_SELF",
			txResult: TemCANNOT_PREAUTH_SELF,
			expected: "temCANNOT_PREAUTH_SELF",
		},
		{
			name:     "TemDST_IS_SRC",
			txResult: TemDST_IS_SRC,
			expected: "temDST_IS_SRC",
		},
		{
			name:     "TemDST_NEEDED",
			txResult: TemDST_NEEDED,
			expected: "temDST_NEEDED",
		},
		{
			name:     "TemINVALID",
			txResult: TemINVALID,
			expected: "temINVALID",
		},
		{
			name:     "TemINVALID_COUNT",
			txResult: TemINVALID_COUNT,
			expected: "temINVALID_COUNT",
		},
		{
			name:     "TemINVALID_FLAG",
			txResult: TemINVALID_FLAG,
			expected: "temINVALID_FLAG",
		},
		{
			name:     "TemMALFORMED",
			txResult: TemMALFORMED,
			expected: "temMALFORMED",
		},
		{
			name:     "TemREDUNDANT",
			txResult: TemREDUNDANT,
			expected: "temREDUNDANT",
		},
		{
			name:     "TemREDUNDANT_SEND_MAX",
			txResult: TemREDUNDANT_SEND_MAX,
			expected: "temREDUNDANT_SEND_MAX",
		},
		{
			name:     "TemRIPPLE_EMPTY",
			txResult: TemRIPPLE_EMPTY,
			expected: "temRIPPLE_EMPTY",
		},
		{
			name:     "TemBAD_WEIGHT",
			txResult: TemBAD_WEIGHT,
			expected: "temBAD_WEIGHT",
		},
		{
			name:     "TemBAD_SIGNER",
			txResult: TemBAD_SIGNER,
			expected: "temBAD_SIGNER",
		},
		{
			name:     "TemBAD_QUORUM",
			txResult: TemBAD_QUORUM,
			expected: "temBAD_QUORUM",
		},
		{
			name:     "TemUNCERTAIN",
			txResult: TemUNCERTAIN,
			expected: "temUNCERTAIN",
		},
		{
			name:     "TemUNKNOWN",
			txResult: TemUNKNOWN,
			expected: "temUNKNOWN",
		},
		{
			name:     "TemDISABLED",
			txResult: TemDISABLED,
			expected: "temDISABLED",
		},

		// Ter codes
		{
			name:     "TerFUNDS_SPENT",
			txResult: TerFUNDS_SPENT,
			expected: "terFUNDS_SPENT",
		},
		{
			name:     "TerINSUF_FEE_B",
			txResult: TerINSUF_FEE_B,
			expected: "terINSUF_FEE_B",
		},
		{
			name:     "TerLAST",
			txResult: TerLAST,
			expected: "terLAST",
		},
		{
			name:     "TerNO_ACCOUNT",
			txResult: TerNO_ACCOUNT,
			expected: "terNO_ACCOUNT",
		},
		{
			name:     "TerNO_AMM",
			txResult: TerNO_AMM,
			expected: "terNO_AMM",
		},
		{
			name:     "TerNO_AUTH",
			txResult: TerNO_AUTH,
			expected: "terNO_AUTH",
		},
		{
			name:     "TerNO_LINE",
			txResult: TerNO_LINE,
			expected: "terNO_LINE",
		},
		{
			name:     "TerNO_RIPPLE",
			txResult: TerNO_RIPPLE,
			expected: "terNO_RIPPLE",
		},
		{
			name:     "TerOWNERS",
			txResult: TerOWNERS,
			expected: "terOWNERS",
		},
		{
			name:     "TerPRE_SEQ",
			txResult: TerPRE_SEQ,
			expected: "terPRE_SEQ",
		},
		{
			name:     "TerPRE_TICKET",
			txResult: TerPRE_TICKET,
			expected: "terPRE_TICKET",
		},
		{
			name:     "TerQUEUED",
			txResult: TerQUEUED,
			expected: "terQUEUED",
		},
		{
			name:     "TerRETRY",
			txResult: TerRETRY,
			expected: "terRETRY",
		},
		{
			name:     "TerSUBMITTED",
			txResult: TerSUBMITTED,
			expected: "terSUBMITTED",
		},
		// Tes codes
		{
			name:     "TesSUCCESS",
			txResult: TesSUCCESS,
			expected: "tesSUCCESS",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.txResult.String()
			require.Equal(t, test.expected, result)
		})
	}
}
