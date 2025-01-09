package models

import (
	"context"
	"fmt"
	"log"
	"time"

	"be20250107/utils/random"

	"github.com/go-redis/redis/v8"
	"gopkg.in/guregu/null.v4"
)

const (
	ProfileStatusSuspended   = "suspended"
	ProfileStatusActive      = "active"
	ProfileStatusDeactivated = "deactivated"

	ContactMediumEmail     = "Email"
	ContactMediumCatalogue = "Catalogue"
	ContactMediumHome      = "Landline"
	ContactMediumLineID    = "LineID"

	ContactTypePrimary     = "Primary"
	ContactTypeAlternative = "Alternative"
	ContactTypeAppleID     = "Apple ID"
	ContactTypeAppleRelay  = "Apple Private Relay"

	BalanceLogActionLockOngoingOrder   = "LOCK_ONGOING_ORDER"
	BalanceLogActionCancelOngoingOrder = "CANCEL_ONGOING_ORDER"
	BalanceLogActionFinishOngoingOrder = "FINISH_ONGOING_ORDER"

	BalanceLogActionRequestWithdraw = "REQUEST_WITHDRAW"
	BalanceLogActionRejectWithdraw  = "REJECT_WITHDRAW"
	BalanceLogActionApproveWithdraw = "APPROVE_WITHDRAW"
	BalanceLogActionRequestTopUp    = "REQUEST_TOPUP"
	BalanceLogActionApproveTopup    = "APPROVE_TOPUP"
	BalanceLogActionRejectTopUp     = "REJECT_TOPUP"

	BalanceLogReferenceTypeOrder             = "ORDER"
	BalanceLogReferenceTypeTopupRequest      = "TOPUP_REQUEST"
	BalanceLogReferenceTypeWithdrawalRequest = "WITHDRAWAL_REQUEST"

	TransactionRequestPending  = "PENDING"
	TransactionRequestApproved = "APPROVED"
	TransactionRequestRejected = "REJECTED"

	RestrictionLevelUnavailable = "UNAVAILABLE"
	RestrictionLevelWarning     = "WARNING"
)

var AvailableContactCombinationLists = []string{
	ContactMediumCatalogue + "-" + ContactTypePrimary,
	ContactMediumCatalogue + "-" + ContactTypeAlternative,
	ContactMediumEmail + "-" + ContactTypePrimary,
	ContactMediumHome + "-" + ContactTypePrimary,
	ContactMediumLineID + "-" + ContactTypePrimary,
}
var TransactionRequestStatuses = []string{TransactionRequestPending, TransactionRequestApproved, TransactionRequestRejected}

type BalanceLogFilter struct {
	From    *time.Time
	To      *time.Time
	Actions []string
	Q       string
}

type VerifiedContactOptions struct {
	Expires   bool
	ExpiresIn time.Duration
}

func GenerateIdempotencyKey() string {
	return random.GenerateString(8, random.NumericCharset+random.UppercaseAlphabeticCharset)
}

type BalanceStatus struct {
	AvailableBalance float64 `json:"available_balance"`
	HoldBalance      float64 `json:"hold_balance"`
	TotalBalance     float64 `json:"total_balance"`
}

type ContactInformation struct {
	Medium             string      `json:"medium"`
	Type               string      `json:"type"`
	VerifiedValue      null.String `json:"verified_value"`
	VerifiedExpiredAt  null.Time   `json:"verified_expired_at"`
	VerifyingValue     null.String `json:"verifying_value"`
	VerifyingExpiresAt null.Time   `json:"verifying_expires_at"`
}

const (
	AccountTypeAdmin  = "admin"
	AccountTypeSystem = "system"
)

func GenerateCode(r *redis.Client, codePrefix string) (string, error) {
	keyCounter := fmt.Sprintf("%v-%v", codePrefix, time.Now().Format("200601"))
	latestCounter := r.Incr(context.Background(), keyCounter)
	if latestCounter.Err() != nil {
		return "", fmt.Errorf("[GenerateCode][Incr]%w", latestCounter.Err())
	}

	_, err := r.Expire(context.Background(), keyCounter, time.Hour*24*32).Result()
	if err != nil {
		log.Printf("[GenerateCode] fail to set expiry on key %v with err:%v", keyCounter, err)
	}
	val, err := latestCounter.Result()
	if err != nil {
		return "", fmt.Errorf("[GenerateCode][Result]%w", err)
	}

	code := fmt.Sprintf("%v-%08v", keyCounter, val)

	return code, nil
}
