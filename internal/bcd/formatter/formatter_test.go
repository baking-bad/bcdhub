package formatter

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"testing"

	"github.com/tidwall/gjson"
)

func TestMichelineToMichelson(t *testing.T) {
	tests := []string{
		"KT18jLSuXycuWi9pL7ByRjqkrWpPF1S6maHY",
		"KT199ibictE9LbQn1kWkhowdiZti6F9mFZQg",
		"KT19VKNKqKDDunvwi4y2r6ugDhCcdhT2zG3c",
		"KT19fth8xoanAobcfdbssDjRTv7So1BCSQt4",
		"KT19kJoPZrPor5yPE5T5rTfLAXRfZVsqzjwT",
		"KT1AJW58kqhEbSMn7w4XVnrRL5zSAP6UrsYQ",
		"KT1At3oM7k94ccMmFCqjAZy42QyaDh2uNqhD",
		"KT1AxGnZL9YzmXZh1Lc4dHhznAeJAQCfML25",
		"KT1BDMQEhMATgVAcwtgqNgZNBM6LEM1PANuM",
		"KT1Ce91KQw3gEEtJiNEagDat2Cr6saM6Cyjm",
		"KT1Cx5ohe4r8QgtP647eidHgZBJhr9L5DSJA",
		"KT1DMBFZ4f6PUHoG3Du5VXEW4XqKfBfoeWJs",
		"KT1DV3mNYRrdpSBzcsJMk9Zd9q43c3iNW9mW",
		"KT1DY8gJ63E7EqpHKLusrDDRahRcEJHwdoNu",
		"KT1E7xh6tvnVMWx7QCZnuWXwcpCJ9UmMWcyK",
		"KT1EUTxJch3jR9VuQ5wV4HeWbs5BnUfQp3N3",
		"KT1EwPxDNyx2y5NSEgKQNLzGSrwBab5Ay8yS",
		"KT1ForWJDps7emzxxdSQpLD9ezdH6Vo7Grzw",
		"KT1FU74GimCeEVRAEZGURb6TWU8jK1N6zFJy",
		"KT1FfZcfsbxXgNKGHpnGWaokXXrvvW1wddGp",
		"KT1FfhRBXiDLuurraaP2u6PkLaPhXSfAdPGY",
		"KT1FkFxTdRGsD2dp6Y1zTRKxtPXqhRJiwQ8L",
		"KT1G2D45tpJ9f1iGVwQHqvupv2syhvMqeWPe",
		"KT1G393LjojNshvMdf68XQD24Hwjn7xarzNe",
		"KT1G72fc8TP3C7WgnaMB8uG3ZbDgfkJNBWEr",
		"KT1GqyAwGGqUbrduNgn4c4aVUXU9UGnXwNmD",
		"KT1HCiJq7ovz5aKqaXft2VwrTBjNsZG18MPH",
		"KT1HRuyp2NLSP9MKBfQyfKoNVUio6Fn7jeDi",
		"KT1J6w61iaASyoTRqM3LHtAWkn9mYvbvm2BT",
		"KT1JcaSFsqJB49R86SfVGo5TsAYzEPni7v1e",
		"KT1JcrtCT2YLiGXNXMMgR63tHTEtg8WNohx3",
		"KT1Jyf1eYy988fGDwmv1EybaaqvmpSH4B9us",
		"KT1K1nH4KWyBV3H37WuRBfpSTq52oE8LGJgh",
		"KT1KVn5cHLPuLoEDmiLEXGfMtNihLtcJtEpM",
		"KT1KXAV7cZmN8ouCqd4rMnMPHy9Wy4Jc3Xvi",
		"KT1Ki9hCRhWERgvVvXvVnFR3ruwM9sR5eLAN",
		"KT1Kmm43Ast1ajWruXsbtra1Eye6sTuriuba",
		"KT1Ktww51i5k2G31DY8aGxvug55sejXTw8Gs",
		"KT1LXmR7aDjTzGLCqaLtqCyZXqoLpUyK2j2n",
		"KT1Lb2v9xrF9994tEZYogrX3og9UHgtRdZBg",
		"KT1LcJ9TriBd42e1MT2CtsZhEfC5tKP8LKnW",
		"KT1LiGjQW3RZeurKpuuaMJyFG1P7Yje1BCb1",
		"KT1ME1G3xGeGdjzfmGYGCWW4FEkTvn88ueZ2",
		"KT1MSZ16hHK9TXdNbZzBUmozAC9yK5snKjoH",
		"KT1MUtNy8bVdwXXaiH2qPAvB6R7Nq9k8BtZg",
		"KT1Md4zkfCvkdqgxAC9tyRYpRUBKmD1owEi2",
		"KT1MqvzsEPoZnbacH18uztqvQdG8x8nKAgFi",
		"KT1Myqcyxp8MNgdB1aAhMpBApZHgVJ634nhm",
		"KT1NAnXMYFDbABvejVCE7TYqMQPr1cUmP25U",
		"KT1NQjN9YgESGZgUm9qSHQL195rgZjpGi7LH",
		"KT1NV5dYRFcd5hf5AQcgo6jodqmYpNc1hiv9",
		"KT1NpCh6tNQDmbmAVbGLxwRBx8jJD4rEFnmC",
		"KT1NwVhTTfmKSYnxecwb6isyJUK5LvuYYDfB",
		"KT1PF3SoynnYGUw3diCjETSbTSEZ1LJMXK9F",
		"KT1Pj9Nn9L13YXoGcCqdXA7r5bL28ghrRi4c",
		"KT1QGqm97QhfRujyZZt67QRaXUnNavjnFngg",
		"KT1QLAVs2iBPDzfGsDXx2CarDUA9yjWXWKgp",
		"KT1QMbabsaiNNeEDD68or5RV7qWZTuguPPdo",
		"KT1QTYM2kcDb6CvAzvC6sYtMGEprqxhoVw4b",
		"KT1QdevirZq7PpMgVFWP6QVRSGbSsdHEUjgt",
		"KT1REci1iEvCLJYYigY3TupRvBugRmdWNWTv",
		"KT1RV1EidsuGckfgLFtDjStSdgbRtyf44jTn",
		"KT1RbgzW6RMPsjhrRRQWZ8fHTLWxW54JyxJm",
		"KT1RmuBSQgU9hnQU9FrweaMoTDcTzG11GmC1",
		"KT1RrfbcDM5eqho4j4u5EbqbaoEFwBsXA434",
		"KT1RvoWEU4MfDBrTbJh4JEAdh3giXtjxkR7k",
		"KT1SawqvsVdAbDzqc4KwPpaS1S1veuFgF9AN",
		"KT1SufMDx6d2tuVe3n6tSYUBNjtV9GgaLgtV",
		"KT1T1QYR6VD2LLtRSP4CHNyKkGbAPHoVu7wc",
		"KT1T6CDRQLRiFU5dszBZWWQwQc8aCbfwX3Mg",
		"KT1T8u994jypfZK68QGAR7rdKRzFHFTXsRDM",
		"KT1TEyRcaJi39jgj4Uuz7VD6Jmn6CDDADv4x",
		"KT1TTDdZqEcVQoPciqLWX5aT9GmfCiW1WDGV",
		"KT1Tc24Zr2G6GrhvoH4n5UM6pJENcWQpvoD8",
		"KT1Tj1P5c2e9q2ow8wcRc76amJj4njLsLpTw",
		"KT1TpKkwKzGwMrWrGnPp9KixhraD2dtE5wE5",
		"KT1Uu2Df4Xn74T3j1cPp34JppjidYsC9yTf5",
		"KT1V29tWwy8B4FBZT2aVpyrzMtBntwj6Hxs4",
		"KT1VgGFh674e2En7fQh7U8WHVyWpwUZB4n4Y",
		"KT1WQAW1sRaykMPYEPpqiL4nrYvdnb8SWTV7",
		"KT1WWy95kwQF3cP8NTCZv1ue7fmBBmAxQpWs",
		"KT1WhouvVKZFH94VXj9pa8v4szvfrBwXoBUj",
		"KT1XFnSFqmXsBmNQVPByuYYkBFoNcXne4Ktu",
		"KT1XRwPmdw7j4LhHgTw8S2dTVbmsXqT6VtpX",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			jsonFile := fmt.Sprintf("./formatter_tests/%v/code_%v.json", tt, tt[:6])

			data, err := ioutil.ReadFile(jsonFile)
			if err != nil {
				t.Error("ioutil.ReadFile code.json error:", err)
			}

			if !gjson.Valid(string(data)) {
				t.Error("invalid json")
			}

			parsedData := gjson.ParseBytes(data)
			result, err := MichelineToMichelson(parsedData, true, DefLineSize)
			if err != nil {
				t.Error("MichelineToMichelson error:", err)
			}

			tzFile := fmt.Sprintf("./formatter_tests/%v/code_%v.tz", tt, tt[:6])
			expected, err := ioutil.ReadFile(tzFile)
			if err != nil {
				t.Error("ioutil.ReadFile code.tz error:", err)
			}

			re := regexp.MustCompile(`\n\s*`)
			exp := re.ReplaceAllString(string(expected), " ")
			exp = strings.ReplaceAll(exp, "{  }", "{}")
			exp = strings.TrimSpace(exp)

			if exp != result {
				t.Errorf("expected != result")
			}
		})
	}
}
