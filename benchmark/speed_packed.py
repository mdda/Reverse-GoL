
class BitField:  # Always surrounded by zeroes
    bitarr = []
    
    def __init_COUNT_BITS():
        #print "__init_COUNT_BITS()"
        arr=[]
        for x in range(0, 1<<9):
            cnt = bin(x).count('1')
            arr.append(cnt)
        return arr
    COUNT_BITS = __init_COUNT_BITS()
    del __init_COUNT_BITS
    
    def __init__(self, n_rows=22, n_cols=22, numeric_array=None):
        self.n_rows=n_rows # Includes the padded region
        self.n_cols=n_cols # Includes the padded region
        self.bitarr = [ 0 for i in range(0, self.n_rows) ]
        
        if numeric_array is None:
            return
            
        r=1 # Start 1 row down
        for arr in numeric_array:  # Each line is an array
            acc=0
            p = 1<<(self.n_cols-1-1)  # Starts 1 column over
            for x in arr:  # Each element of the array
                if x: 
                    acc |= p
                p >>= 1
            self.bitarr[r] |= acc
            r=r+1
            
    def pretty_print(self):
        spacer = '-' * self.n_cols
        for row in self.bitarr:
            b = bin(row).lstrip('-0b')
            print (spacer+b)[-self.n_cols:].replace('0','-')

    def iterate(self):
        ## These are constants - the game bits pass over them
        tb_filter  = 7
        mid_filter = 5
        current_filter = 2 
            
        arr_new=[0]
        for r in range(1, self.n_rows-1):
            r_top = self.bitarr[r-1]
            r_mid = self.bitarr[r]
            r_bot = self.bitarr[r+1]
            
            acc = 0
            p = 2 # Start in the middle row, one column in
            
            for c in range(1, self.n_cols-1):
                cnt = self.COUNT_BITS[
                  ((r_top & tb_filter) << 6 )|
                  ((r_mid & mid_filter) << 3 )|
                   (r_bot & tb_filter)
                ]
            
                #if True: acc |= p  # Check bit-twiddling bounds
                
                # Return next state according to the game rules:
                #  exactly 3 neighbors: on,
                #  exactly 2 neighbors: maintain current state,
                #  otherwise: off.
                #  return alive == 3 || alive == 2 && f.Alive(x, y)
                if (cnt==3) or (cnt==2 and (r_mid & current_filter)) : 
                    acc |= p  
                
                # Move the 'setting-bit' over
                p <<= 1
                
                # Shift the arrays over into base filterable position
                r_top >>= 1
                r_mid >>= 1
                r_bot >>= 1
            
            #print "Appending %d" % (acc)
            arr_new.append(acc)
        arr_new.append(0)
        self.bitarr = arr_new
    
    #def copy(self):
        

glider = [[0,0,1],
          [1,0,1],
          [0,1,1]]

z = BitField(numeric_array=glider)

print 'Initial state:'
z.pretty_print()

for i in range(65):
    z.iterate()
   
print 'Final state:'
z.pretty_print()


def test_timing():
    import timeit
    
    def time_iter():
        z = BitField(numeric_array=glider)
        for i in range(65):
            z.iterate()
        
    t=timeit.Timer(time_iter)
    print t.repeat(1, 1000)
    
test_timing()

