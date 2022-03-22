#!/bin/sh

# Enable debugging if variable is set
test $DEBUG_SHELL -eq 0 || set -o xtrace

# Usage: publishTopic [OPTIONS]
# Options:
#	-h host
#	-t topic
#	-m message
#	-q qos
#	-r retain
#	-s enable DTLS ("s" stands for "secure")
#	-p predefinedTopic
publishTopic() {
	local host="" topic="" msg="" qos="" retain="" dtls="" predefinedTopic=""
	while getopts "h:t:m:q:rsp:" arg; do
		case "$arg" in
		h)
			host="$OPTARG"
			;;
		t)
			topic="$OPTARG"
			;;
		m)
			msg="$OPTARG"
			;;
		q)
			qos="$OPTARG"
			;;
		r)
			retain="1"
			;;
		s)
			dtls="1"
			;;
		p)
			predefinedTopic="$OPTARG"
			;;
		esac
	done
	shift $((OPTIND - 1))

	bisquitt-pub \
		--host "$host" \
		--topic "$topic" \
		--message "$msg" \
		$(test -z "$qos" || echo --qos "$qos") \
		$(test -z "$retain" || echo --retain) \
		$(test -z "$dtls" || echo --dtls --self-signed --insecure) \
		$(test -z "$predefinedTopic" || echo --predefined-topic "$predefinedTopic")
}

# Usage: subscribeTopic [OPTIONS]
# Options:
#	-h host
#	-t topic
#	-f file path to file where client's output will be written
#	-q qos
#	-s enable DTLS ("s" stands for "secure")
#	-p predefinedTopic
subscribeTopic() {
	local host="" topic="" file="" pidfile="" qos="" dtls="" predefinedTopic=""
	while getopts "h:t:f:q:sp:" arg; do
		case "$arg" in
		h)
			host="$OPTARG"
			;;
		t)
			topic="$OPTARG"
			;;
		f)
			file="$OPTARG"
			pidfile="$file.pid"
			;;
		q)
			qos="$OPTARG"
			;;
		s)
			dtls="1"
			;;
		p)
			predefinedTopic="$OPTARG"
			;;
		esac
	done
	shift $((OPTIND - 1))

	bisquitt-sub \
		--host "$host" \
		--topic "$topic" \
		$(test -z "$qos" || echo --qos "$qos") \
		$(test -z "$dtls" || echo --dtls --self-signed --insecure) \
		$(test -z "$predefinedTopic" || echo --predefined-topic "$predefinedTopic") >"$file"&
	echo $! >"$pidfile"
}

# Usage: unsubscribeTopic [OPTIONS]
# Options:
#	-f file path to file where client's output will be written
unsubscribeTopic() {
	local file="" pidfile=""
	while getopts "f:" arg; do
		case "$arg" in
		f)
			file="$OPTARG"
			pidfile="$file.pid"
			;;
		esac
	done
	shift $((OPTIND - 1))

	if [ ! -f "$pidfile" ]; then
		echo "subscriber is not running"
		return
	fi
	kill "$(cat $pidfile)"
}

# Usage: makeMessage <num>
makeMessage() {
	printf "Hello World %d" $1
}

# Usage: testMessagePassthrough [OPTIONS]
# Options:
#	-h host
#	-i subscribeTopic
#	-o publishTopic
#	-q qos
#	-s enable DTLS ("s" stands for "secure")
#	-p predefinedTopic
testMessagePassthrough() {
	local numMessages=20
	local output="$(mktemp)"
	local host="" subscribeTopic="" publishTopic="" qos="" dtls=""
	local predefinedTopic=""
	while getopts "h:i:o:q:sp:" arg; do
		case "$arg" in
		h)
			host="$OPTARG"
			;;
		i)
			subscribeTopic="$OPTARG"
			;;
		o)
			publishTopic="$OPTARG"
			;;
		q)
			qos="$OPTARG"
			;;
		s)
			dtls="1"
			;;
		p)
			predefinedTopic="$OPTARG"
			;;
		esac
	done
	shift $((OPTIND - 1))

	subscribeTopic \
		-h "$host" \
		-t "$subscribeTopic" \
		-f "$output" \
		$(test -z "$qos" || test "$qos" -eq 3 || echo -q "$qos") \
		$(test -z "$dtls" || echo -s) \
		$(test -z "$predefinedTopic" || echo -p "$predefinedTopic")

	# Wait before blasting messages.
	sleep 3

	for i in $(seq 1 $numMessages); do
		local message="$(makeMessage $i)"
		publishTopic \
			-h "$host" \
			-t "$publishTopic" \
			-m "$message" \
			$(test -z "$qos" || echo -q "$qos") \
			$(test -z "$dtls" || echo -s) \
			$(test -z "$predefinedTopic" || echo -p "$predefinedTopic")
	done

	unsubscribeTopic -f "$output"

	local receivedMessages=0
	while read -r line; do
		local message="$(makeMessage $(($receivedMessages + 1)))"
		echo "$line" | grep "$topic: $message" >/dev/null
		if [ $? -eq 0 ]; then
			receivedMessages=$(($receivedMessages + 1))
		fi
	done <"$output"

	if [ $receivedMessages -ne $numMessages ]; then
		echo "not enough messages received ($receivedMessages/$numMessages)"
		exit 1
	fi
}

# Usage: testMessagePassthroughQOS0 [OPTIONS]
# Options:
#	-h host
#	-s enable DTLS ("s" stands for "secure")
testMessagePassthroughQOS0() {
	local topic="messages-qos0"
	local qos=0
	local host="" dtls=""
	while getopts "h:s" arg; do
		case "$arg" in
		h)
			host="$OPTARG"
			;;
		s)
			dtls="1"
			;;
		esac
	done
	shift $((OPTIND - 1))

	testMessagePassthrough \
		-h "$host" \
		-i "$topic" \
		-o "$topic" \
		-q "$qos" \
		$(test -z "$dtls" || echo -s)
}

# Usage: testMessagePassthroughQOS1 [OPTIONS]
# Options:
#	-h host
#	-s enable DTLS ("s" stands for "secure")
testMessagePassthroughQOS1() {
	local topic="messages-qos1"
	local qos=1
	local host="" dtls=""
	while getopts "h:s" arg; do
		case "$arg" in
		h)
			host="$OPTARG"
			;;
		s)
			dtls="1"
			;;
		esac
	done
	shift $((OPTIND - 1))

	testMessagePassthrough \
		-h "$host" \
		-i "$topic" \
		-o "$topic" \
		-q "$qos" \
		$(test -z "$dtls" || echo -s)
}

# Usage: testMessagePassthroughQOS2 [OPTIONS]
# Options:
#	-h host
#	-s enable DTLS ("s" stands for "secure")
testMessagePassthroughQOS2() {
	local topic="messages-qos2"
	local qos=2
	local host="" dtls=""
	while getopts "h:s" arg; do
		case "$arg" in
		h)
			host="$OPTARG"
			;;
		s)
			dtls="1"
			;;
		esac
	done
	shift $((OPTIND - 1))

	testMessagePassthrough \
		-h "$host" \
		-i "$topic" \
		-o "$topic" \
		-q "$qos" \
		$(test -z "$dtls" || echo -s)
}

# Usage: testMessagePassthroughQOS3 [OPTIONS]
# Options:
#	-h host
#	-s enable DTLS ("s" stands for "secure")
testMessagePassthroughQOS3() {
	local topic="messages-predefined-qos3"
	local topicID=18
	local qos=3
	local host="" dtls=""
	while getopts "h:s" arg; do
		case "$arg" in
		h)
			host="$OPTARG"
			;;
		s)
			dtls="1"
			;;
		esac
	done
	shift $((OPTIND - 1))

	testMessagePassthrough \
		-h "$host" \
		-i "$topic" \
		-o "$topic" \
		-q "$qos" \
		-p "$topic;$topicID" \
		$(test -z "$dtls" || echo -s)
}

# Usage: testMessagePassthroughWildcardTopic [OPTIONS]
# Options:
#	-h host
#	-s enable DTLS ("s" stands for "secure")
testMessagePassthroughWildcardTopic() {
	local topicSubscribed="messages-wildcard/+/data/#"
	local topicPublished="messages-wildcard/device/data/all"
	local host="" dtls=""
	while getopts "h:s" arg; do
		case "$arg" in
		h)
			host="$OPTARG"
			;;
		s)
			dtls="1"
			;;
		esac
	done
	shift $((OPTIND - 1))

	testMessagePassthrough \
		-h "$host" \
		-i "$topicSubscribed" \
		-o "$topicPublished" \
		$(test -z "$dtls" || echo -s)
}

# Usage: testMessagePassthroughShortTopic [OPTIONS]
# Options:
#	-h host
#	-s enable DTLS ("s" stands for "secure")
testMessagePassthroughShortTopic() {
	local topic="ab"
	local host="" dtls=""
	while getopts "h:s" arg; do
		case "$arg" in
		h)
			host="$OPTARG"
			;;
		s)
			dtls="1"
			;;
		esac
	done
	shift $((OPTIND - 1))

	testMessagePassthrough \
		-h "$host" \
		-i "$topic" \
		-o "$topic" \
		$(test -z "$dtls" || echo -s)
}

# Usage: testMessagePassthroughShortTopic [OPTIONS]
# Options:
#	-h host
#	-s enable DTLS ("s" stands for "secure")
testMessagePassthroughPredefinedTopic() {
	local topic="messages-predefined"
	local topicID=17
	local host="" dtls=""
	while getopts "h:s" arg; do
		case "$arg" in
		h)
			host="$OPTARG"
			;;
		s)
			dtls="1"
			;;
		esac
	done
	shift $((OPTIND - 1))

	testMessagePassthrough \
		-h "$host" \
		-i "$topic" \
		-o "$topic" \
		-p "$topic;$topicID" \
		$(test -z "$dtls" || echo -s)
}

# Usage: testMessageRetain [OPTIONS]
# Options:
#	-h host
#	-s enable DTLS ("s" stands for "secure")
testMessageRetain() {
	local topic="messages-retain"
	local numMessages=10
	local output="$(mktemp)"
	local host="" dtls=""
	while getopts "h:s" arg; do
		case "$arg" in
		h)
			host="$OPTARG"
			;;
		s)
			dtls="1"
			;;
		esac
	done
	shift $((OPTIND - 1))

	for i in $(seq 1 $numMessages); do
		local message="$(makeMessage $i)"
		publishTopic \
			-h "$host" \
			-t "$topic" \
			-m "$message" \
			-r \
			$(test -z "$dtls" || echo -s)
	done

	subscribeTopic \
		-h "$host" \
		-t "$topic" \
		-f "$output" \
		$(test -z "$dtls" || echo -s)

	# Wait before blasting messages.
	sleep 3

	unsubscribeTopic -f "$output"

	local receivedMessages=0
	while read -r line; do
		local message="$(makeMessage $numMessages)"
		echo "$line" | grep "$topic: $message \[retained\]" >/dev/null
		if [ $? -eq 0 ]; then
			receivedMessages=$(($receivedMessages + 1))
		fi
	done <"$output"

	if [ $receivedMessages -ne 1 ]; then
		echo "not enough messages received ($receivedMessages/1)"
		exit 1
	fi
}

# Usage: startTest <function>
startTest() {
	local host="bisquitt"
	local f="$1"
	shift 1

	echo "Running $f"
	msg="$(eval $f -h "$host" $@)"
	if [ $? -ne 0 ]
	then
		echo "$msg [ERROR]"
		exit 1
	else
		echo "$msg [OK]"
	fi
}

# Usage: startTestDTLS <function>
startTestDTLS() {
	local host="bisquitt-dtls"
	local f="$1"
	shift 1

	echo "Running $f (DTLS)"
	msg="$(eval $f -h "$host" -s $@)"
	if [ $? -ne 0 ]
	then
		echo "[ERROR]: $msg"
		exit 1
	else
		echo "$msg [OK]"
	fi
}

startTestDTLS "testMessagePassthroughWildcardTopic"
startTestDTLS "testMessagePassthroughShortTopic"
startTestDTLS "testMessagePassthroughPredefinedTopic"
startTestDTLS "testMessageRetain"
startTestDTLS "testMessagePassthroughQOS0"
startTestDTLS "testMessagePassthroughQOS1"
startTestDTLS "testMessagePassthroughQOS2"
startTestDTLS "testMessagePassthroughQOS3"

startTest "testMessagePassthroughWildcardTopic"
startTest "testMessagePassthroughShortTopic"
startTest "testMessagePassthroughPredefinedTopic"
startTest "testMessageRetain"
startTest "testMessagePassthroughQOS0"
startTest "testMessagePassthroughQOS1"
startTest "testMessagePassthroughQOS2"
startTest "testMessagePassthroughQOS3"
