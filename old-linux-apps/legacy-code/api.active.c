
/* RescueTime::Network::API::activate(std::__cxx11::basic_string<wchar_t, std::char_traits<wchar_t>,
   std::allocator<wchar_t> > const&, std::__cxx11::basic_string<wchar_t, std::char_traits<wchar_t>,
   std::allocator<wchar_t> > const&, std::__cxx11::basic_string<wchar_t, std::char_traits<wchar_t>,
   std::allocator<wchar_t> > const&) */

API * __thiscall
RescueTime::Network::API::activate
          (API *this,basic_string *param_1,basic_string *param_2,basic_string *param_3)

{
  size_t sVar1;
  long in_FS_OFFSET;
  undefined *local_68 [2];
  undefined local_58 [24];
  long local_40;

  local_40 = *(long *)(in_FS_OFFSET + 0x28);
  local_68[0] = local_58;
  sVar1 = wcslen(L"[API::activate]");
  std::__cxx11::basic_string<>::_M_construct<>
            ((wchar_t *)local_68,L"[API::activate]",(int)sVar1 * 4 + 0xa1e820);
                    /* try { // try from 00631658 to 0063165c has its CatchHandler @ 0063170b */
  rt::debug_log((basic_string *)local_68,0xc0);
  if (local_68[0] != local_58) {
    operator.delete(local_68[0]);
  }
  local_68[0] = local_58;
  sVar1 = wcslen(L"");
  std::__cxx11::basic_string<>::_M_construct<>((wchar_t *)local_68,L"",(int)sVar1 * 4 + 0x9f06ec);
                    /* try { // try from 006316b0 to 006316b4 has its CatchHandler @ 006316ee */
  perform_activate((basic_string *)this,param_1,param_2,param_3);
  if (local_68[0] != local_58) {
    operator.delete(local_68[0]);
  }
  if (local_40 == *(long *)(in_FS_OFFSET + 0x28)) {
    return this;
  }
                    /* WARNING: Subroutine does not return */
  __stack_chk_fail();
}