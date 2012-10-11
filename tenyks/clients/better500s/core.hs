{-# LANGUAGE OverloadedStrings #-}
import Control.Monad (forever)
import Data.Maybe
import Database.Redis.Redis

host = "127.0.0.1"
port = "6379" 
channel = "tenyks.services.broadcast_to"

getMessage :: (Message String) -> String
getMessage (MMessage s1 s2) = s2
getMessage _ = ""

main = do
    putStrLn host    
    putStrLn port

    db <- connect host port
    subscribe db [channel] :: IO [Message ()]
    forever $ do
        message <- listen db 1000
        if isJust message
        then do
            let msg = getMessage $ fromJust message
            putStrLn msg
        else do
            return ()    
    disconnect db

    return ()
