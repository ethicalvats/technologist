package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(Command interface{}) (index, Term, isleader)
//   start agreement on a new log entry
// rf.GetState() (Term, isLeader)
//   ask a Raft for its current Term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	"errors"
	"lib-raft/labrpc"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// import "bytes"
// import "../labgob"

const min = 10
const max = 30
const ELECTIONTIMEOUT = 5

//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in Lab 3 you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh; at that point you can add fields to
// ApplyMsg, but set CommandValid to false for these other uses.
//
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
}

type LogEntry struct {
	Command interface{}
	Term    int
}

//
// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu        sync.Mutex     // Lock to protect shared access to this peer's state
	peers     []*labrpc.Node // RPC end points of all peers
	persister *Persister     // Object to hold this peer's persisted state
	me        *labrpc.Node   // this peer's index into peers[]
	dead      int32          // set by Kill()

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	currentTerm         int
	isLeader            bool
	isFollower          bool
	isCandidate         bool
	votedFor            int
	nextElectionTimeOut time.Duration
	electionResetTime   time.Time
	log                 []LogEntry
	commitIndex         int
	lastApplied         int
	nextIndex           map[int]int
	matchIndex          map[int]int
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) getState() (int, bool) {

	var term int
	var isLeader bool
	// Your code here (2A).
	rf.mu.Lock()
	defer rf.mu.Unlock()
	term = rf.currentTerm
	isLeader = rf.isLeader
	return term, isLeader
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}

//
// example RequestVote RPC arguments structure.
// field names must start with capital letters!
//
type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

type AppendEntriesArgs struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

//
// example RequestVote RPC reply structure.
// field names must start with capital letters!
//
type RequestVoteReply struct {
	// Your data here (2A).
	Term        int
	VoteGranted bool
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

//
// example RequestVote RPC handler.
//
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) error {
	// Your code here (2A, 2B).
	requesterTerm := args.Term
	rf.mu.Lock()
	defer rf.mu.Unlock()
	DPrintf("%d voting for %d", rf.me, args.CandidateId)
	//if rf.isFollower {
	if requesterTerm < rf.currentTerm {
		reply.VoteGranted = false
	} else {
		DPrintf("%d voted for %v", rf.me, rf.votedFor)
		if rf.votedFor == -1 {
			reply.VoteGranted = true
			rf.currentTerm = requesterTerm
			rf.isCandidate = false
			rf.isLeader = false
			rf.isFollower = true
		}
	}
	reply.Term = rf.currentTerm
	//}
	now := time.Now()
	electionReset := rf.electionResetTime
	if now.After(electionReset) {
		randT := (((rand.Intn(max-min+1) + min) * ELECTIONTIMEOUT) + ELECTIONTIMEOUT) * 100
		rf.nextElectionTimeOut = time.Duration(randT) * time.Millisecond //* time.Duration(rf.me+1)
		rf.electionResetTime = time.Now()
	}
	return nil
}

// If followers command for leader commitIndex does not match return false
// Leader should reduce its current index for the follower and send another AE
// till leader finds a matching index and then send AE will all the prev logs
// follower should delete its logs from this point onwards and accept leaders logs

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) error {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	var followerPrevLog LogEntry
	var followerPrevTerm int

	followerLogLen := len(rf.log)
	followerPrevIndex := followerLogLen - 1

	if followerLogLen > 0 {
		followerPrevLog = rf.log[followerPrevIndex]
		followerPrevTerm = followerPrevLog.Term
	}

	//DPrintf("%d term %d received AE from %d args %v", rf.me, rf.currentTerm, args.LeaderId, args)
	if rf.currentTerm <= args.Term {

		//DPrintf("%d follower prev index %v prev term %v log len %v", rf.me, followerPrevIndex, followerPrevTerm, followerLogLen)

		if followerLogLen > args.PrevLogIndex && args.PrevLogIndex > -1 {
			followerLogTerm := rf.log[args.PrevLogIndex].Term

			if followerLogTerm != args.PrevLogTerm {
				DPrintf("%d cleaning logs till index %d", rf.me, args.PrevLogIndex)
				correctedLogs := rf.log[:args.PrevLogIndex]
				rf.log = correctedLogs
			}
		}

		if (followerPrevIndex == args.PrevLogIndex &&
			followerPrevTerm == args.PrevLogTerm) ||
			followerLogLen == 0 {
			rf.currentTerm = args.Term
			reply.Success = true
			rf.isFollower = true
			rf.isCandidate = false
			rf.isLeader = false
			for _, entry := range args.Entries {
				rf.log = append(rf.log, entry)
			}
			//DPrintf("%d is follower %t", rf.me, rf.isFollower)
		} else {
			reply.Success = false
		}

		if followerLogLen < args.LeaderCommit {
			rf.commitIndex = followerPrevIndex
		} else {
			rf.commitIndex = args.LeaderCommit
		}
	} else if followerPrevIndex+1 <= args.LeaderCommit {
		rf.currentTerm = args.Term
		rf.isFollower = true
		rf.isCandidate = false
		rf.isLeader = false
	}
	reply.Term = rf.currentTerm
	now := time.Now()
	electionReset := rf.electionResetTime
	if now.After(electionReset) {
		randT := (((rand.Intn(max-min+1) + min) * ELECTIONTIMEOUT) + ELECTIONTIMEOUT) * 100
		rf.nextElectionTimeOut = time.Duration(randT) * time.Millisecond //* time.Duration(rf.me+1)
		rf.electionResetTime = time.Now()
	}
	//DPrintf("%d log %v", rf.me, rf.log)
	//DPrintf("%d reply %v", rf.me, reply)
	//DPrintf("%d election timeout %v", rf.me, rf.nextElectionTimeOut)
	return nil
}

//
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
//
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	DPrintf("%d sendRequestVote %d", rf.me, server)
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	//DPrintf("%d sendAppendEntries %d", rf.me, server)
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next Command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// Command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the Command will appear at
// if it's ever committed. the second return value is the current
// Term. the third return value is true if this server believes it is
// the leader.
//

type StartReply struct {
	IsLeader bool
	Index    int
	Term     int
}

type StartArgs struct {
	Data interface{}
}

func (rf *Raft) Start(command *StartArgs, rep *StartReply) error {
	term := -1
	isLeader := true
	commitIndex := 0

	// Your code here (2B).
	DPrintf("\033[1;33m %d received client command %v \033[0m\n ", rf.me, command)
	rf.mu.Lock()
	isLeader = rf.isLeader
	term = rf.currentTerm
	rf.mu.Unlock()

	if rf.isLeader {
		var newEntries []LogEntry
		newSyncIndex := 0
		newLog := LogEntry{Term: term, Command: command}
		rf.mu.Lock()
		rf.log = append(rf.log, newLog)
		//if rf.commitIndex > 0 {
		//	newEntries = rf.log[rf.commitIndex:]
		//}else{
		//	newEntries = append(newEntries, newLog)
		//}
		newEntries = append(newEntries, newLog)
		rf.mu.Unlock()
		_, prevLogIndex, prevLogTerm := rf.findPrevLogEntry()
		DPrintf("leader log %v", rf.log)
		consensus := 1
		wg := sync.WaitGroup{}
		for p := range rf.peers {
			if p != rf.me.Id {
				wg.Add(1)
				go func(peer int, entries []LogEntry) {
					rf.mu.Lock()
					isLeader = rf.isLeader
					rf.mu.Unlock()

					if isLeader {
						args := AppendEntriesArgs{
							Term:         rf.currentTerm,
							LeaderId:     rf.me.Id,
							PrevLogIndex: prevLogIndex,
							PrevLogTerm:  prevLogTerm,
							LeaderCommit: rf.commitIndex,
							Entries:      entries,
						}
						reply := AppendEntriesReply{}
						//DPrintf("%d sending AE to %d", rf.me, peer)
						rf.sendAppendEntries(peer, &args, &reply)

						if reply.Success {
							rf.mu.Lock()
							rf.matchIndex[peer] = prevLogIndex
							rf.nextIndex[peer] = prevLogIndex + 1
							rf.mu.Unlock()
							newSyncIndex = prevLogIndex + 1
							consensus++
							DPrintf("%d peer reply %t consensus %d", peer, reply.Success, consensus)
						} else {
							rf.mu.Lock()
							leaderTerm := rf.currentTerm
							if reply.Term > leaderTerm {
								DPrintf("%d received higher term %d no longer leader", rf.me, reply.Term)
								rf.currentTerm = reply.Term
								rf.isLeader = false
								rf.isFollower = true
								rf.isCandidate = false
								rf.votedFor = -1
								rf.electionResetTime = time.Now().Add(10 * time.Second) // push lost leaders election reset by 10s
							}
							rf.mu.Unlock()
							if prevLogIndex > -1 && reply.Term <= leaderTerm {
								sync, _ := rf.synWithFollower(peer, prevLogIndex)
								newSyncIndex = prevLogIndex + 1
								if sync {
									consensus++
								}
							}
						}
					}
					wg.Done()
				}(p, newEntries)
			}
		}
		wg.Wait()
		if consensus > len(rf.peers)/2 {
			rf.mu.Lock()
			DPrintf("new sync index %d", newSyncIndex)
			if newSyncIndex == 0 {
				rf.commitIndex++
				commitIndex = rf.commitIndex
			} else {
				for newSyncIndex > rf.commitIndex {
					rf.commitIndex++
					commitIndex = rf.commitIndex
					DPrintf("leader committed index %d with consensus %d", commitIndex, consensus)
					time.Sleep(10 * time.Millisecond)
				}
			}
			rf.mu.Unlock()
		} else {
			commitIndex = rf.commitIndex + 1
		}
	}
	rf.mu.Lock()
	isLeader = rf.isLeader
	rf.mu.Unlock()

	if isLeader {
		DPrintf("\033[1;32m %d return index %d term %d is leader %t \033[0m", rf.me, commitIndex, term, isLeader)
	}
	rep.IsLeader = isLeader
	rep.Index = commitIndex
	rep.Term = term

	return nil
}

func (rf *Raft) synWithFollower(follower int, followerPrevIndex int) (bool, int) {
	var isLeader bool
	for i := followerPrevIndex; i >= 0; i-- {
		rf.mu.Lock()
		isLeader = rf.isLeader
		rf.mu.Unlock()

		if isLeader {
			DPrintf("%d failed AE syncing follower %d at index %d", rf.me, follower, i)
			leaderTerm := rf.log[i].Term
			args := AppendEntriesArgs{
				Term:         rf.currentTerm,
				LeaderId:     rf.me.Id,
				PrevLogIndex: i,
				PrevLogTerm:  leaderTerm,
			}
			reply := AppendEntriesReply{}
			DPrintf("%d sending AE to %d", rf.me, follower)
			rf.sendAppendEntries(follower, &args, &reply)
			if reply.Success {
				DPrintf("found matching index %d", i)
				rf.mu.Lock()
				logEntries := rf.log[i+1:]
				rf.mu.Unlock()
				//logEntries = append(logEntries, log)
				DPrintf("sending entries %v", logEntries)
				args.Entries = logEntries
				rf.sendAppendEntries(follower, &args, &reply)
				return true, len(logEntries)
			}
		}
	}
	return false, 0
}

func (rf *Raft) findPrevLogEntry() (error, int, int) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	logLen := len(rf.log)
	if logLen > 1 {
		prevEntry := rf.log[logLen-1]
		//DPrintf("%d index %v", rf.me, logLen-1)
		return nil, logLen - 1, prevEntry.Term
	} else {
		return errors.New("no logs"), -1, 0
	}

}

//
// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
//
func (rf *Raft) kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//

/*
	Modify Make() to create a background goroutine that will
	1. kick off leader election periodically
	2. by sending out RequestVote RPCs when it hasn't heard from another peer for a while.

	This way
	a peer will learn who is the leader, if there is already a leader, or become the leader itself.
	Implement the RequestVote() RPC handler so that servers will vote for one another.
*/

/*
	When servers start up, they begin as followers. A server remains in follower state as long as it
	receives valid RPCs from a leader or candidate. Leaders send periodic heartbeats
	(AppendEntries RPCs that carry no log entries) to all followers in order to maintain their authority.
	If a follower receives no communication over a period of time called the election timeout,
	then it assumes there is no viable leader and begins an election to choose a new leader.
*/

func (rf *Raft) startElection() {
	rf.mu.Lock()
	rf.isLeader = false
	rf.isFollower = false
	rf.isCandidate = true //becomes candidate from follower after starting the election
	rf.currentTerm++
	voteCount := 1
	candidatePeers := rf.peers
	candidate := rf.me
	candidateTerm := rf.currentTerm
	DPrintf("%d starting election with Term %d", rf.me, rf.currentTerm)
	rf.mu.Unlock()

	wg := sync.WaitGroup{}

	for p := range candidatePeers {
		if p != candidate.Id {
			wg.Add(1)
			go func(peer int, self int, term int) {
				rf.mu.Lock()
				args := RequestVoteArgs{
					Term:         term,
					CandidateId:  self,
					LastLogIndex: -1,
					LastLogTerm:  0,
				}
				reply := RequestVoteReply{}
				rf.mu.Unlock()
				rf.sendRequestVote(peer, &args, &reply)
				DPrintf("%d vote %d args %v reply %v", rf.me, peer, args, reply)
				if reply.VoteGranted {
					voteCount++
				}
				wg.Done()
			}(p, candidate.Id, candidateTerm)
		}
	}

	wg.Wait()
	DPrintf("%d election vote count %d", rf.me, voteCount)
	rf.processResults(voteCount, candidate.Id)
}

func (rf *Raft) initializeLeaderEntries() {
	nextIndex := make(map[int]int, 2)
	rf.nextIndex = nextIndex
	rf.nextIndex[0] = 0
	rf.nextIndex[1] = 0

	matchIndex := make(map[int]int, 2)
	rf.matchIndex = matchIndex
	rf.matchIndex[0] = 0
	rf.matchIndex[1] = 0

}

func (rf *Raft) processResults(voteCount int, candidate int) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	maxVote := len(rf.peers)/2 + 1
	if voteCount >= maxVote {
		DPrintf("%d is leader", rf.me)
		rf.isLeader = true
		rf.isCandidate = false
		rf.isFollower = false
		rf.initializeLeaderEntries()
		go func() {
			for {
				rf.beginLeadership(candidate)
				time.Sleep(20 * time.Millisecond)
			}
		}()
	} else {
		rf.isCandidate = false
		rf.isFollower = true
		rf.electionResetTime = time.Now().Add(10 * time.Second)
	}
	return
}

func (rf *Raft) beginLeadership(candidate int) {
	_, prevLogIndex, prevLogTerm := rf.findPrevLogEntry()
	for p := range rf.peers {
		if p != candidate && rf.isLeader {
			var entries []LogEntry
			args := AppendEntriesArgs{
				Term:         rf.currentTerm,
				LeaderId:     rf.me.Id,
				PrevLogIndex: prevLogIndex,
				PrevLogTerm:  prevLogTerm,
				LeaderCommit: rf.commitIndex,
				Entries:      entries,
			}
			reply := AppendEntriesReply{}
			go rf.sendAppendEntries(p, &args, &reply)
			//DPrintf("%d HB to %d args %v reply %v", rf.me, p, args, reply)
			if reply.Term > rf.currentTerm {
				rf.isLeader = false
			}
		}
	}
}

func (rf *Raft) checkTimeout() {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	var elapsed time.Duration
	var timeoutDuration time.Duration
	var electionResetTime time.Time
	var isFollower bool

	for {
		<-ticker.C
		rf.mu.Lock()
		timeoutDuration = rf.nextElectionTimeOut
		electionResetTime = rf.electionResetTime
		isFollower = rf.isFollower
		rf.mu.Unlock()
		elapsed = time.Since(electionResetTime)
		if elapsed >= timeoutDuration {
			if isFollower {
				DPrintf("%d my leader not alive after %v starting election", rf.me, elapsed)
				go rf.startElection()
			}
			//return
		}
		//rf.mu.Unlock()

		//if !rf.nextElectionTimeOut.IsZero() &&
		//	time.Now().After(rf.nextElectionTimeOut) {
		//	if rf.isFollower { //TODO Find appropriate randomization function to give enough gap between two elections
		//		DPrintf("%d my leader not alive staring election", rf.me)
		//		go rf.StartElection()
		//	}
		//}
	}
}

func (rf *Raft) checkNewLogEntry(applyCh chan ApplyMsg) {
	for {
		rf.mu.Lock()
		commitIndex := rf.commitIndex
		lastApplied := rf.lastApplied
		rf.mu.Unlock()
		if lastApplied < commitIndex {
			DPrintf("%d commitIndex %d lastApplied %d", rf.me, commitIndex, lastApplied)
			rf.mu.Lock()
			appliedLog := rf.lastApplied
			rf.lastApplied += 1
			newLogs := rf.log[appliedLog:commitIndex]
			DPrintf("%d new logs %v", rf.me, newLogs)
			rf.mu.Unlock()
			for index, log := range newLogs {
				msg := ApplyMsg{CommandValid: true, Command: log.Command, CommandIndex: appliedLog + index + 1} // +1 for commitIndex
				//DPrintf("%d new log entry %v", rf.me, msg)
				applyCh <- msg
			}
			//msg := ApplyMsg{CommandValid: true, Command: newLog.Command, CommandIndex: commitIndex} // +1 for commitIndex
			////DPrintf("%d new log entry %v", rf.me, msg)
			//applyCh <- msg
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func Make(peers []*labrpc.Node, me *labrpc.Node,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (2A, 2B, 2C).
	DPrintf("Make server %d", me)

	rf.currentTerm = 0
	rf.isLeader = false
	rf.isFollower = true // starts as follower
	rf.isCandidate = false
	rf.votedFor = -1
	rf.commitIndex = 0
	rf.lastApplied = 0
	rf.log = []LogEntry{}

	//randInt := rand.Intn(MAX - MIN) + MIN
	randInt := (((rand.Intn(max-min+1) + min) * ELECTIONTIMEOUT) + ELECTIONTIMEOUT) * 100
	//time.Sleep(time.Duration(randInt) * 1000 * time.Millisecond)
	startTime := time.Now()
	nextTimeout := time.Millisecond * time.Duration(randInt) // * time.Duration(rf.me+1)
	DPrintf("start time %v timeout %v", startTime, startTime.Add(nextTimeout))
	rf.electionResetTime = startTime
	rf.nextElectionTimeOut = nextTimeout
	//if rf.me == 1 { //TODO Not sure if as an initial condition it is Ok?
	//	go rf.StartElection()
	//}

	go rf.checkTimeout()
	go rf.checkNewLogEntry(applyCh)

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	return rf
}
