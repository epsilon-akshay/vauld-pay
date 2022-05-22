package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"net/http"
)

type TranferReq struct {
	EmailID         string  `json:"emailID"`
	ToUserName      string  `json:"toUserName"`
	Amount          float64 `json:"amount"`
	ToFiatCurrency  string  `json:"fiatCurrency"`
	ToCyptoCurrency string  `json:"crytoCurrency"`
}

func Transfer(s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var t TranferReq
		err := decoder.Decode(&t)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println(err)
			s := fmt.Sprintf("{success:false, err:%s}", err.Error())
			w.Write([]byte(s))
			return
		}

		key := t.EmailID
		_, err = s.Client.Get(context.Background(), key).Result()
		if err == redis.Nil {
			w.WriteHeader(http.StatusInternalServerError)
			s := fmt.Sprintf(`{"success"":"false"", "err"": "please do your kyc %s"}`, err.Error())
			w.Write([]byte(s))
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s := fmt.Sprintf(`{"success"":"false"", "err"": "something went wrong %s"}`, err.Error())
			w.Write([]byte(s))
			return
		}

		amountInUSDT := ConvertUSDTFromFiat(t.Amount)

		err = AddToWallet(t.ToUserName, t.EmailID, amountInUSDT, s.Client, t.ToCyptoCurrency, t.ToFiatCurrency)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println(err)
			s := fmt.Sprintf("{success:false, err:%s}", err.Error())
			w.Write([]byte(s))
			return
		}

		err = ConvertCurrency(amountInUSDT, t.ToUserName, t.EmailID, s.Client, t.ToCyptoCurrency, t.ToFiatCurrency)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println("2", err)
			s := fmt.Sprintf(`{"success"":"false"", "err"": "something went wrong %s"}`, err.Error())
			w.Write([]byte(s))
			return
		}

		w.WriteHeader(http.StatusOK)
		l := fmt.Sprintf("{success:true}")
		w.Write([]byte(l))
		return
	}
}
func ConvertCurrency(amount float64, userName string, email string, client *redis.Client, cryptoCurr string, fiatCurr string) error {
	fiatAmount := amount * 75
	fmt.Println("asdasd", fiatAmount)

	r, err := client.Get(context.Background(), email).Bytes()
	if err == redis.Nil {
		return fmt.Errorf("couldnt find kyc address")
	} else if err != nil {
		return fmt.Errorf("couldnt call redis")
	}
	var val Value
	err = json.Unmarshal(r, &val)
	if err != nil {
		return err
	}

	var wAddrs []WalletValue
	for _, i := range val.WalletAdd {
		t := i

		if i.Username == userName {
			if cryptoCurr != "" {
				final := []Amount{}
				for _, j := range i.CryptoAmountDetails {
					am := j.Val
					if cryptoCurr == j.Currency {
						am = j.Val - amount
					}
					final = append(final, Amount{Val: am, Currency: j.Currency})
				}
				t.CryptoAmountDetails = final
			}

			if fiatCurr != "" {
				final := []Amount{}
				for _, j := range i.FiatAmountDetails {
					am := j.Val
					if fiatCurr == j.Currency {
						fmt.Println("qweqwe", j.Val, j.Currency)
						am = j.Val + fiatAmount
					}
					final = append(final, Amount{Val: am, Currency: j.Currency})
				}
				t.FiatAmountDetails = final
			}
		}
		wAddrs = append(wAddrs, t)
	}

	v := Value{
		Name:      val.Name,
		Id:        val.Id,
		WalletAdd: wAddrs,
	}

	err = client.Set(context.Background(), email, v, 0).Err()
	if err != nil {
		return err
	}

	return nil
}
func AddToWallet(userName string, email string, amount float64, client *redis.Client, cryptoCurr string, fiatCurr string) error {
	r, err := client.Get(context.Background(), email).Bytes()
	if err == redis.Nil {
		return fmt.Errorf("couldnt find kyc address")
	} else if err != nil {
		return fmt.Errorf("couldnt call redis")
	}
	var val Value
	err = json.Unmarshal(r, &val)
	if err != nil {
		fmt.Println("asdsadad", err)
		return err
	}

	var wAddrs []WalletValue
	for _, i := range val.WalletAdd {
		t := i

		if i.Username == userName {
			if cryptoCurr != "" {
				final := []Amount{}
				for _, j := range i.CryptoAmountDetails {
					am := j.Val
					if cryptoCurr == j.Currency {
						am = j.Val + amount
					}
					final = append(final, Amount{Val: am, Currency: j.Currency})
				}
				t.CryptoAmountDetails = final
			}

			//if fiatCurr != "" {
			//	final := []Amount{}
			//	for _, j := range i.FiatAmountDetails {
			//		am := j.Val
			//		if fiatCurr == j.Currency {
			//			am = j.Val + amount
			//		}
			//		final = append(final, Amount{Val: am, Currency: j.Currency})
			//	}
			//	t.FiatAmountDetails = final
			//}
		}

		wAddrs = append(wAddrs, t)
	}

	v := Value{
		Name:      val.Name,
		Id:        val.Id,
		WalletAdd: wAddrs,
	}

	err = client.Set(context.Background(), email, v, 0).Err()
	if err != nil {
		return err
	}

	return nil
}
func ConvertUSDTFromFiat(amount float64) float64 {
	return amount * 1
}
