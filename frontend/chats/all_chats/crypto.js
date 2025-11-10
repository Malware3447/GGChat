const cryptoUtils = {
    rsaKeyParams: {
        name: "RSA-OAEP",
        modulusLength: 2048,
        publicExponent: new Uint8Array([0x01, 0x00, 0x01]),
        hash: "SHA-256",
    },
    aesKeyParams: {
        name: "AES-GCM",
        length: 256,
    },
    encoder: new TextEncoder(),
    decoder: new TextDecoder(),

    arrayBufferToBase64(buffer) {
        let binary = '';
        const bytes = new Uint8Array(buffer);
        const len = bytes.byteLength;
        for (let i = 0; i < len; i++) {
            binary += String.fromCharCode(bytes[i]);
        }
        return window.btoa(binary);
    },
    base64ToArrayBuffer(base64) {
        const binary_string = window.atob(base64);
        const len = binary_string.length;
        const bytes = new Uint8Array(len);
        for (let i = 0; i < len; i++) {
            bytes[i] = binary_string.charCodeAt(i);
        }
        return bytes.buffer;
    },
    async exportKeyToPEM(key, type = "public") {
        const exported = await window.crypto.subtle.exportKey(
            type === "public" ? "spki" : "pkcs8",
            key
        );
        const b64 = this.arrayBufferToBase64(exported);
        const header = type === "public" ? "PUBLIC KEY" : "PRIVATE KEY";
        return `-----BEGIN ${header}-----\n${b64}\n-----END ${header}-----`;
    },
    async importKeyFromPEM(pem, type = "public") {
        const b64 = pem.replace(/-----(BEGIN|END) (PUBLIC|PRIVATE) KEY-----/g, "").trim();
        const buffer = this.base64ToArrayBuffer(b64);
        const format = type === "public" ? "spki" : "pkcs8";
        const usage = type === "public" ? "encrypt" : "decrypt";
        
        return window.crypto.subtle.importKey(
            format,
            buffer,
            this.rsaKeyParams,
            true,
            [usage]
        );
    },


    /**
     * 1. Инициализация: Проверяет localStorage, генерирует ключи, если их нет.
     * @returns {Object} { privateKey: CryptoKey, publicKeyPEM: string }
     */
    async initKeys() {
        let privateKey = null;
        let publicKey = null;
        let publicKeyPEM = null;

        const storedPrivateKeyPEM = localStorage.getItem("user_private_key");

        if (storedPrivateKeyPEM) {
            privateKey = await this.importKeyFromPEM(storedPrivateKeyPEM, "private");
            const storedPublicKeyPEM = localStorage.getItem("user_public_key");
            publicKeyPEM = storedPublicKeyPEM;

        } else {
            const keyPair = await window.crypto.subtle.generateKey(
                this.rsaKeyParams,
                true,
                ["encrypt", "decrypt"]
            );
            privateKey = keyPair.privateKey;
            publicKey = keyPair.publicKey;

            const privateKeyPEM = await this.exportKeyToPEM(privateKey, "private");
            publicKeyPEM = await this.exportKeyToPEM(publicKey, "public");
            
            localStorage.setItem("user_private_key", privateKeyPEM);
            localStorage.setItem("user_public_key", publicKeyPEM);
            
            console.log("Новая пара RSA-ключей сгенерирована и сохранена.");
        }
        
        return { privateKey, publicKeyPEM };
    },

    /**
     * 2. Шифрование сообщения (Гибридная схема)
     * @param {string} plainText - Текст сообщения
     * @param {Map<number, string>} publicKeysMap - Map, где { user_id -> public_key_pem }
     * @returns {Object} { content: string (base64), keys: Object }
     */
    async encryptMessage(plainText, publicKeysMap) {
        const aesKey = await window.crypto.subtle.generateKey(
            this.aesKeyParams,
            true,
            ["encrypt", "decrypt"]
        );
        const aesKeyExported = await window.crypto.subtle.exportKey("raw", aesKey);
        
        const iv = window.crypto.getRandomValues(new Uint8Array(12));
        const encodedText = this.encoder.encode(plainText);
        
        const encryptedContentBuffer = await window.crypto.subtle.encrypt(
            { name: "AES-GCM", iv: iv },
            aesKey,
            encodedText
        );
        
        const ivAndContent = new Uint8Array(iv.length + encryptedContentBuffer.byteLength);
        ivAndContent.set(iv);
        ivAndContent.set(new Uint8Array(encryptedContentBuffer), iv.length);
        const contentB64 = this.arrayBufferToBase64(ivAndContent.buffer);

        const keysMap = {};
        
        for (const [userId, publicKeyPEM] of publicKeysMap.entries()) {
            if (!publicKeyPEM) continue;
            
            const rsaPublicKey = await this.importKeyFromPEM(publicKeyPEM, "public");
            
            const encryptedAesKeyBuffer = await window.crypto.subtle.encrypt(
                this.rsaKeyParams,
                rsaPublicKey,
                aesKeyExported
            );
            
            keysMap[userId] = this.arrayBufferToBase64(encryptedAesKeyBuffer);
        }
        
        return {
            content: contentB64, 
            keys: keysMap 
        };
    },

    /**
     * 3. Расшифровка сообщения (Гибридная схема)
     * @param {string} contentB64 - Зашифрованный контент (IV + Сообщение) из B64
     * @param {string} encryptedKeyB64 - *Твой* зашифрованный AES-ключ из B64
     * @param {CryptoKey} privateKey - *Твой* импортированный приватный ключ
     * @returns {string} plainText - Расшифрованный текст
     */
    async decryptMessage(contentB64, encryptedKeyB64, privateKey) {
        if (!contentB64 || !encryptedKeyB64 || !privateKey) {
            throw new Error("Отсутствуют данные для расшифровки.");
        }

        const encryptedKeyBuffer = this.base64ToArrayBuffer(encryptedKeyB64);
        const aesKeyBuffer = await window.crypto.subtle.decrypt(
            this.rsaKeyParams,
            privateKey,
            encryptedKeyBuffer
        );
        
        const aesKey = await window.crypto.subtle.importKey(
            "raw",
            aesKeyBuffer,
            this.aesKeyParams,
            true,
            ["encrypt", "decrypt"]
        );

        const ivAndContent = this.base64ToArrayBuffer(contentB64);
        const iv = ivAndContent.slice(0, 12);
        const contentBuffer = ivAndContent.slice(12);

        const decryptedBuffer = await window.crypto.subtle.decrypt(
            { name: "AES-GCM", iv: iv },
            aesKey,
            contentBuffer
        );

        return this.decoder.decode(decryptedBuffer);
    }
};