import numpy

def iterate(Z):
    # find number of neighbours that each square has
    N = numpy.zeros(Z.shape)
    N[1:, 1:] += Z[:-1, :-1]
    N[1:, :-1] += Z[:-1, 1:]
    N[:-1, 1:] += Z[1:, :-1]
    N[:-1, :-1] += Z[1:, 1:]
    N[:-1, :] += Z[1:, :]
    N[1:, :] += Z[:-1, :]
    N[:, :-1] += Z[:, 1:]
    N[:, 1:] += Z[:, :-1]
    # a live cell is killed if it has fewer 
    # than 2 or more than 3 neighbours.
    part1 = ((Z == 1) & (N < 4) & (N > 1)) 
    # a new cell forms if a square has exactly three members
    part2 = ((Z == 0) & (N == 3))
    return (part1 | part2).astype(int)

Z = numpy.array([[0,0,0,0,0,0],
                 [0,0,0,1,0,0],
                 [0,1,0,1,0,0],
                 [0,0,1,1,0,0],
                 [0,0,0,0,0,0],
                 [0,0,0,0,0,0]])

glider = numpy.array([[0,0,1],
                 [1,0,1],
                 [0,1,1]])

Z = numpy.zeros((22,22), dtype=numpy.int)
Z[1:1+glider.shape[0], 1:1+glider.shape[1]] = glider

print 'Initial state:'
print Z[1:-1,1:-1]
for i in range(65):
    Z = iterate(Z)
print 'Final state:'
#print Z[1:-1,1:-1]
print Z[:,:]

print "Problem with edges..."


def test_timing():
    import timeit
    
    def time_iter():
        Z = numpy.zeros((22,22), dtype=numpy.int)
        Z[1:1+glider.shape[0], 1:1+glider.shape[1]] = glider

        for i in range(65):
            Z = iterate(Z)

    t=timeit.Timer(time_iter)
    print t.repeat(1, 1000)
    
test_timing()
