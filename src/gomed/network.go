package gomed

/*
Copyright 2012 Jeremy Sullivan

This file is part of gomed.

gomed is free software: you can redistribute it and/or modify
it under the terms of the GNU Lesser General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

gomed is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Lesser General Public License for more details.

You should have received a copy of the GNU Lesser General Public License
along with gomed.  If not, see <http://www.gnu.org/licenses/>.
*/

import (
  "fmt"
  //"strings"
  //"encoding/xml"
  //"os"
  //"strconv"
  "bytes"
  "net"
  //"xmlx"
  //"io"
)

func Send(addy string, msg string) {

  address, reserr := net.ResolveTCPAddr("tcp", addy)
  if reserr != nil{
    println(reserr)
  }

  conn, err := net.DialTCP("tcp", nil, address)
  if err != nil {
    println(err)
  }
  
  n, cerr := conn.Write([]byte(msg))
  if cerr != nil {
    println(cerr)
  }
  println(n)

  conn.Close()
}

func StartSimpleServer() {
  //TODO: HL7 handle function as argument?
  //Start a simple server with no config
  //Useful for testing
  l, lerr := net.Listen("tcp", ":9090")
  if lerr != nil {
    println(lerr)
  }

  //a first loop to accept connections
  for {
    conn, err := l.Accept()
    if err != nil {
      println(err)
    }
    //handle those connections concurrently
    go func(c net.Conn) {
      //a second loop to read
      for {
        //Be sure the buffer is init'd this way.
        //Obviously, this isn't going to handle more then a meg gracefully
        b := make([]byte, 1024)
        n, nerr := conn.Read(b)
        if nerr != nil {
          fmt.Printf("read error: %v\n", nerr)
        } else {
          fmt.Printf("main read %v bytes\n", n)
          fmt.Printf("those bytes: %v\n", string(b))
          //ends the client if !quit is sent at the begging of the msg
          //ignores anything after.
          //should probably have if bytes.HasSuffix or Contains
          if bytes.HasPrefix(b, []byte("!quit")){
            println("Closing the connection...")
            c.Close()
            break
          }
        }
      }
     }(conn)

  }
}

