"""simple testinfra test file"""


def test_passwd_file(host):
    """validate passwd file"""
    passwd = host.file('/etc/passwd')
    assert passwd.contains('root')
    assert passwd.user == 'root'
    assert passwd.group == 'root'
    assert passwd.mode == 0o644
