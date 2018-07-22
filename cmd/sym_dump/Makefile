# Makefile to check for errors in sym_dump output.

H=$(shell find . -type f -name '*.h')
H_OBJ=$(H:.h=.o)

C=$(shell find . -type f -name '*.c')
C_OBJ=$(C:.c=.o)

CPP=$(shell find . -type f -name '*.cpp')
CPP_OBJ=$(CPP:.cpp=.o)

all: $(H_OBJ) $(C_OBJ) $(CPP_OBJ)

%.o: %.h
	clang -c -I ./ -Wno-return-type -Wno-main-return-type -Wno-incompatible-library-redeclaration -o $@ $<

%.o: %.c
	clang -c -I ./ -Wno-return-type -Wno-main-return-type -Wno-incompatible-library-redeclaration -o $@ $<

%.o: %.cpp
	clang -x c -c -I ./ -Wno-return-type -Wno-main-return-type -Wno-incompatible-library-redeclaration -o $@ $<

clean:
	$(RM) -v $(H_OBJ) $(C_OBJ) $(CPP_OBJ)
