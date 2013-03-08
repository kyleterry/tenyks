{-# LANGUAGE OverloadedStrings, ScopedTypeVariables #-}
import Control.Monad (forever)
import Data.Maybe

import Database.Redis.Redis
import Reactive.Banana
import Text.JSON

host :: String
host = "127.0.0.1"

port :: String
port = "6379"

channel :: String
channel = "tenyks.services.broadcast_to"

getMessage :: (Message String) -> String
getMessage (MMessage _ s2) = s2
getMessage _ = ""

main :: IO ()
main = do
    putStrLn host
    putStrLn port

    db <- connect host port
    _ <- subscribe db [channel] :: IO [Message ()]
    _ <- forever $ do
        message <- listen db 1000
        if isJust message
        then do
            let msg = getMessage $ fromJust message
            putStrLn msg
        else do
            return ()
    disconnect db

    return ()
