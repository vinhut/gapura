
package utils
  
import (  
   "crypto/aes"  
   "crypto/cipher"
   "encoding/hex"
   )

func GCM_encrypt(key string, plaintext string, iv []byte, additionalData []byte) string {  
   block, err := aes.NewCipher([]byte(key))  
   if err != nil {  
      panic(err.Error())  
   }  
   aesgcm, err := cipher.NewGCM(block)  
   if err != nil {  
      panic(err.Error())  
   }  
   ciphertext := aesgcm.Seal(nil, iv, []byte(plaintext), additionalData)
   stringed := hex.EncodeToString(iv) + "-" + hex.EncodeToString(ciphertext)
   return stringed  
}  
  
func GCM_decrypt(key string,ct string,iv string, additionalData []byte) (string,error) {  
   ciphertext, _ := hex.DecodeString(ct)  
   iv_decode,_ := hex.DecodeString(iv)
   block, err := aes.NewCipher([]byte(key))  
   if err != nil {  
      return "error",err  
   }  
   aesgcm, err := cipher.NewGCM(block)  
   if err != nil {  
      return "error",err  
   }  
   plaintext, err := aesgcm.Open(nil, iv_decode, ciphertext, additionalData)  
   if err != nil {  
      return "error",err  
   }  
   s := string(plaintext[:])  
   return s,nil  
}