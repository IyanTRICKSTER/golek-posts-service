db.createUser({
    user: "ayocodedb",
    pwd: "secret",
    roles: [
        { role: "readWrite", db: "golek_posts" }
    ],
    mechanisms: ["SCRAM-SHA-1"],
})