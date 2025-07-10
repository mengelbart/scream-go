.PHONY: clean

TARGET=scream

$(TARGET): libscream.a
	go build ./...

libscream.a: scream/code/ScreamTx.o scream/code/ScreamV2TxStream.o scream/code/ScreamV2Tx.o ScreamTxC.o scream/code/ScreamRx.o ScreamRxC.o
	ar r $@ $^

%.o: %.cpp
	g++ -Wno-overflow -Wno-write-strings -O2 -o $@ -c $^

clean:
	rm -f *.o *.so *.a $(TARGET)
