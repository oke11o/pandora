#!/usr/bin/env bash

set -e
# shellcheck disable=SC2155
export _SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

# shellcheck source=_functions.sh
source "$_SCRIPT_DIR/_functions.sh"

# shellcheck source=_variables.sh
source "$_SCRIPT_DIR/_variables.sh"

should_run=0
should_check=0


_DIRS=()
while [[ $# -gt 0 ]]; do
    case "$1" in
    --help | -h)
        echo "Usage: $(basename "$0") TEST_DIR1 [TEST_DIR2]..."
        echo ""
        echo "Sequentially run and check results of tests defined in directories passed as arguments,"
        echo "print summary."
        echo ""
        echo "Specifically, for each provided argument:"
        echo "1. call '_test_run.sh TEST_DIR' to run the test (see '_test_run.sh --help')"
        echo "2. call '_test_check.sh --id TEST_ID --dir TEST_DIR' to check the results (see '_test_check.sh --help')"
        exit 0
        ;;
    --should_run)
        should_run=1
        shift
        ;;
    --should_check)
        should_check=1
        shift
        ;;
    *)
        _DIRS+=("$1")
        shift
        ;;
    esac
done
[ $(($should_run+$should_check)) -eq 0 ] && _log "At least one of --should_run, --should_check flags must be specified" && exit 1
assert_installed yc jq curl

_logv 1 "YC CLI profile: ${VAR_CLI_PROFILE:-"current aka <$(yc_ config profile list | grep ' ACTIVE')>"}"
_logv 1 ""
_log "Got dirs for tests: ${_DIRS[*]}"
_log -f <<EOF
Execution:
|--------------------
| folder: ${VAR_FOLDER_ID:-$(yc_ config get folder-id)}
| skip results check: $VAR_SKIP_TEST_CHECK
|
| data bucket: $VAR_DATA_BUCKET
| extra test labels: $VAR_TEST_EXTRA_LABELS
| extra test description: $VAR_TEST_EXTRA_DESCRIPTION
|
| output local dir: $VAR_OUTPUT_DIR
|--------------------
EOF
_log ""

if [[ -z "${VAR_FOLDER_ID:-$(yc_ config get folder-id)}" ]]; then
    _log "Folder ID must be specified either via YC_LT_FOLDER_ID or via CLI profile."
    exit 1
fi

declare -i _tests_total="${#_DIRS[@]}"
declare -i _tests_failed=0
declare _tests_failure_reports=()

_log_push_stage ""
_log_push_stage ""
_test_ids=
_test_num=-1
for _test_dir in "${_DIRS[@]}"; do
    _log_pop_stage
    _log_pop_stage
    _log_push_stage "[$_test_dir]"
    _log_push_stage "[ENTER]"
    _test_num=$((_test_num+1))
    _log "Checking..."
    [[ -z "$_test_dir" ]] && echo "variable _test_dir is empty. exit" && exit 1
    if [[ ! -d "$_test_dir" ]]; then
        _msg="FAILED: test dir does not exist"
        _tests_failure_reports+=("$(_log "$_msg" 2> >(tee /dev/stderr))")
        _tests_failed=$((_tests_failed + 1))
        continue
    fi

    _test_id=
    if [ $should_run -eq 1 ]; then
        _log_stage "[RUN]"
        _log "Running..."

        if _test=$(run_script "$_SCRIPT_DIR/_test_run.sh" "$_test_dir"); then
            _test_id=$(jq -r '.id' <<< "$_test")
            _logv 1 "ID=$_test_id"
            _logv 1 "STATUS=$(jq -r '.summary.status' <<< "$_test")"
            _log "FINISHED: $(yc_test_url "$_test_id")/test-report)"
        else
            _msg="FAILED: error; test=$_test"
            _tests_failure_reports+=("$(_log "$_msg" 2> >(tee /dev/stderr))")
            _tests_failed=$((_tests_failed + 1))
            continue
        fi
        _test_ids="${_test_ids} ${_test_id}"
        test_ids=$_test_ids
    fi

    if [ $should_check -eq 1 ]; then
      # TODO make separated scripts
        test_ids_arr=($test_ids)
        _test_id=${test_ids_arr[$_test_num]}
#        _test_id=${_test_id:-"$test_id"}
        _log_stage "[CHECK]"
        _log "test_id is $_test_id"
        [[ -z "$_test_id" ]] && echo "test_id is empty. Continue" && continue
        if [[ "${VAR_SKIP_TEST_CHECK:-0}" == 0 ]]; then
            _out_dir="$VAR_OUTPUT_DIR/$_test_dir"
            _resfile="$_out_dir/check_result.txt"
            mkdir -p "$_out_dir"

            _log "Performing checks..."
            if run_script "$_SCRIPT_DIR/_test_check.sh" --id "$_test_id" --dir "$_test_dir" -o "$_out_dir"  >"$_resfile"; then
                _logv 1 -f <"$_resfile"
                _log "ALL CHECKS PASSED"
            else
                _log -f <"$_resfile"
                _msg="FAILED: checks did not pass. Result in $_resfile"
                _tests_failure_reports+=("$(_log "$_msg" 2> >(tee /dev/stderr))")
                _tests_failed=$((_tests_failed + 1))
            fi
        else
            _log "skipped due to YC_LT_SKIP_TEST_CHECK"
        fi

    fi
    _log ""
done

_log_pop_stage
_log_pop_stage
_log_stage ""

_summary_header="[ OK - $((_tests_total - _tests_failed)) | FAILED - $_tests_failed ]"
_log "==================== $_summary_header ===================="
_log ""
if ((_tests_failed != 0)); then
    _log "$_tests_failed out of $_tests_total tests have failed:"
    for _msg in "${_tests_failure_reports[@]}"; do
        _log "$_msg"
    done
fi
_log ""
_log "==================== $_summary_header ===================="

#echo "$_tests_failed"
echo "$_test_ids"

#exit "$_tests_failed"
